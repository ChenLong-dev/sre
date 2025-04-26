package httpclient

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/base/hook"
	"gitlab.shanhai.int/sre/library/base/runtime"
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	K8sHostSuffix = ".svc.cluster.local"

	LoadBalancerModeNormal = iota
	LoadBalancerModeCluster
)

type LoadBalancer struct {
	logger hook.Logger

	balancer sync.Map

	k8sClient *kubernetes.Clientset

	globalContext context.Context

	mode int
}

type HostBalancer struct {
	host        string
	serviceName string
	env         string

	once    sync.Once
	mutex   sync.Mutex
	isWatch int32

	endpoints []string
	idx       int64
}

func (b *HostBalancer) IsWatch() bool {
	return atomic.LoadInt32(&b.isWatch) == 1
}

func (b *HostBalancer) readIndex() int {
	if len(b.endpoints) == 0 {
		return 0
	}
	v := atomic.AddInt64(&b.idx, 1) % int64(len(b.endpoints))
	atomic.StoreInt64(&b.idx, v)
	return int(v)
}

func NewLoadBalancer(ctx context.Context, logger hook.Logger) (*LoadBalancer, error) {
	balancer := &LoadBalancer{
		globalContext: ctx,
		logger:        logger,
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		if err == rest.ErrNotInCluster {
			balancer.mode = LoadBalancerModeNormal
		} else {
			return nil, err
		}
	} else {
		balancer.mode = LoadBalancerModeCluster
		balancer.k8sClient, err = kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}
	}

	return balancer, nil
}

func (lb *LoadBalancer) watch(ctx context.Context, host string) error {
	value, ok := lb.balancer.Load(host)
	if !ok {
		return errors.Errorf("%s isn't add", host)
	}
	balancer := value.(*HostBalancer)

	if balancer.serviceName == "" || balancer.env == "" {
		return nil
	}

	if atomic.LoadInt32(&balancer.isWatch) == 1 {
		return errors.Errorf("%s already watch", host)
	}

	balancer.mutex.Lock()
	defer balancer.mutex.Unlock()
	if balancer.isWatch == 1 {
		return errors.Errorf("%s already watch", host)
	}

	watchItf, err := lb.k8sClient.CoreV1().
		Endpoints(balancer.env).
		Watch(ctx, metaV1.ListOptions{
			Watch:         true,
			FieldSelector: fields.OneTermEqualSelector("metadata.name", balancer.serviceName).String(),
		})
	if err != nil {
		return err
	}
	defer atomic.StoreInt32(&balancer.isWatch, 1)

	go func() {
		defer func() {
			watchItf.Stop()
			balancer.endpoints = make([]string, 0)
			atomic.StoreInt64(&balancer.idx, 0)
			atomic.StoreInt32(&balancer.isWatch, 0)
		}()

		watchChan := watchItf.ResultChan()
		for {
			select {
			case event := <-watchChan:
				switch event.Type {
				case watch.Added, watch.Modified, watch.Deleted:
					endpoint := event.Object.(*v1.Endpoints)
					if len(endpoint.Subsets) == 0 {
						break
					}
					subset := endpoint.Subsets[0]
					if len(subset.Ports) == 0 {
						break
					}
					port := subset.Ports[0]

					endpoints := make([]string, len(subset.Addresses))
					for i, address := range subset.Addresses {
						endpoints[i] = fmt.Sprintf("%s:%d", address.IP, port.Port)
					}

					balancer.endpoints = endpoints

					lb.logger.Print(map[string]interface{}{
						"start_time": time.Now(),
						"source":     runtime.GetDefaultFilterCallers(),
						"extra_message": fmt.Sprintf(
							"update endpoints: name:%s env:%s endpoint:%#v",
							balancer.serviceName, balancer.env, balancer.endpoints,
						),
					})
				case "":
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (lb *LoadBalancer) MustGetEndpoint(ctx context.Context, host string) string {
	endpoint, _ := lb.GetEndpoint(ctx, host)
	if endpoint == "" {
		return host
	}

	return endpoint
}

func (lb *LoadBalancer) GetEndpoint(ctx context.Context, host string) (string, error) {
	value, ok := lb.balancer.Load(host)
	if !ok {
		return "", errors.Errorf("%s isn't add", host)
	}
	balancer := value.(*HostBalancer)

	if l := len(balancer.endpoints); l == 0 {
		return "", nil
	} else {
		return balancer.endpoints[balancer.readIndex()], nil
	}
}

func (lb *LoadBalancer) Watch(ctx context.Context, host string) error {
	if lb.mode != LoadBalancerModeCluster {
		return errors.New("load balancer is not cluster mode")
	}

	err := lb.watch(lb.globalContext, host)
	if err != nil {
		return err
	}

	return nil
}

func (lb *LoadBalancer) IsAdded(ctx context.Context, host string) bool {
	_, ok := lb.balancer.Load(host)
	return ok
}

func (lb *LoadBalancer) Add(ctx context.Context, host string) error {
	_, ok := lb.balancer.Load(host)
	if ok {
		return errors.Errorf("%s already add", host)
	}

	balancer := &HostBalancer{
		host: host,
		idx:  0,
	}

	svcName, env, err := ParseNameAndEnvFromK8sHost(host)
	if err == nil {
		balancer.serviceName = svcName
		balancer.env = env
	}
	lb.balancer.Store(host, balancer)

	err = lb.Watch(lb.globalContext, host)
	if err != nil {
		return err
	}

	return nil
}

func ParseNameAndEnvFromK8sHost(host string) (string, string, error) {
	body := strings.TrimSuffix(host, K8sHostSuffix)
	if len(body) == len(host) {
		return "", "", errors.Errorf("%s isn't k8s host", host)
	}

	res := strings.Split(body, ".")
	if len(res) == 2 {
		return res[0], res[1], nil
	} else if len(res) == 3 {
		// dev环境
		return res[1], res[2], nil
	} else {
		return "", "", errors.Errorf("%s host is invalid", host)
	}
}
