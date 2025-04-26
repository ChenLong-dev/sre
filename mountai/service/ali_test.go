package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"rulai/models/entity"
	"rulai/models/req"
	"rulai/models/resp"
)

func Test_Service_AliPrivateZoneRecordOperations(t *testing.T) {
	ctx := context.Background()

	t.Run("阿里云私有域记录流程测试", func(t *testing.T) {
		ts := time.Now().UnixNano()
		domainName := fmt.Sprintf("ams-unittest-ali-privatezone-%d", ts)
		ip := "1.2.3.5"
		updater := fmt.Sprintf("ams-单元测试-阿里云私有域操作-%d", ts)
		domainController := "单元测试假用户-001"

		err := s.createAliPrivateZoneRecord(ctx, domainName, ip, updater, domainController, entity.ARecord)
		assert.Nil(t, err)

		// 重复错误忽略测试
		err = s.createAliPrivateZoneRecord(ctx, domainName, ip, updater, domainController, entity.ARecord)
		assert.Nil(t, err)

		records, err := s.getAliPrivateZoneRecord(ctx, &req.GetQDNSRecordsReq{
			DomainType:       entity.ARecord,
			DomainRecordName: domainName,
			DomainValue:      ip,
			PrivateZone:      AliK8sPrivateZone,
			PageNumber:       1,
			PageSize:         req.GetQDNSRecordsPageSizeLimit,
		})
		assert.Nil(t, err)
		if assert.Len(t, records, 1) {
			assert.Equal(t, domainController, records[0].DomainController)
			assert.Equal(t, AliK8sPrivateZone, records[0].DomainName)
			assert.Equal(t, domainName, records[0].Name)
			assert.Equal(t, entity.ARecord, records[0].Type)
			assert.Equal(t, ip, records[0].Value)
		}

		anotherDomainName := fmt.Sprintf("ams-unittest-ali-privatezone-another-%d", ts)
		err = s.createAliPrivateZoneRecord(ctx, anotherDomainName, ip, updater, domainController, entity.ARecord)
		assert.Nil(t, err)

		records, err = s.getAliPrivateZoneRecord(ctx, &req.GetQDNSRecordsReq{
			DomainType:       entity.ARecord,
			DomainRecordName: anotherDomainName,
			DomainValue:      ip,
			PrivateZone:      AliK8sPrivateZone,
			PageNumber:       1,
			PageSize:         req.GetQDNSRecordsPageSizeLimit,
		})
		assert.Nil(t, err)
		if assert.Len(t, records, 2) {
			for i := range records {
				assert.Equal(t, domainController, records[i].DomainController)
				assert.Equal(t, AliK8sPrivateZone, records[i].DomainName)
				assert.Equal(t, entity.ARecord, records[i].Type)
				assert.Equal(t, ip, records[i].Value)
			}

			assert.Equal(t, domainName, records[0].Name)
			assert.Equal(t, anotherDomainName, records[1].Name)
		}

		err = s.deleteAliPrivateZoneRecord(ctx, domainName, ip, updater, entity.ARecord)
		assert.Nil(t, err)

		records, err = s.getAliPrivateZoneRecord(ctx, &req.GetQDNSRecordsReq{
			DomainType:       entity.ARecord,
			DomainRecordName: domainName,
			DomainValue:      ip,
			PrivateZone:      AliK8sPrivateZone,
			PageNumber:       1,
			PageSize:         req.GetQDNSRecordsPageSizeLimit,
		})
		assert.Nil(t, err)
		if assert.Len(t, records, 1) {
			assert.Equal(t, domainController, records[0].DomainController)
			assert.Equal(t, AliK8sPrivateZone, records[0].DomainName)
			assert.Equal(t, anotherDomainName, records[0].Name)
			assert.Equal(t, entity.ARecord, records[0].Type)
			assert.Equal(t, ip, records[0].Value)
		}

		err = s.deleteAliPrivateZoneRecord(ctx, anotherDomainName, ip, updater, entity.ARecord)
		assert.Nil(t, err)

		records, err = s.getAliPrivateZoneRecord(ctx, &req.GetQDNSRecordsReq{
			DomainType:       entity.ARecord,
			DomainRecordName: anotherDomainName,
			DomainValue:      ip,
			PrivateZone:      AliK8sPrivateZone,
			PageNumber:       1,
			PageSize:         req.GetQDNSRecordsPageSizeLimit,
		})
		assert.Nil(t, err)
		assert.Len(t, records, 0)
	})
}

func TestService_createPrivateZoneRecordEntry(t *testing.T) {
	type args struct {
		ctx     context.Context
		project *resp.ProjectDetailResp
		app     *resp.AppDetailResp
		task    *resp.TaskDetailResp
	}
	_, appDetailResp, taskDetailResp := getPATByProjectID(1)

	appDetailResp.ServiceName = "istio-ingressgateway"
	taskDetailResp.Namespace = "istio-system"
	taskDetailResp.EnvName = entity.AppEnvStg

	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
		{"1", args{app: appDetailResp, project: &resp.ProjectDetailResp{}, task: taskDetailResp}, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, SVC.createPrivateZoneRecordEntry(tt.args.ctx, tt.args.project, tt.args.app, tt.args.task),
				fmt.Sprintf("createPrivateZoneRecordEntry(%v, %v, %v, %v)", tt.args.ctx, tt.args.project,
					tt.args.app, tt.args.task))
		})
	}
}
