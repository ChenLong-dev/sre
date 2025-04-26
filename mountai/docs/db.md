## mongo

### 1.1 project 的资源列表

| field        | type      | desc                        |
| ------------ | --------- | --------------------------- |
| \_id         | object id |                             |
| team_id      | string    | 团队 id                     |
| project_id   | string    | 项目 id                     |
| rds          | [string]  | mysql                       |
| redis        | [string]  | redis                       |
| mongo        | [string]  | mongo                       |
| hbase        | [string]  | hbase                       |
| cdn          | [string]  | cdn                         |
| ecs          | [string]  | ecs                         |
| env          | [string]  | 环境                         |
| commit_id    | string    | git config-center commit id |
| updated_time | datetime  |                             |
| created_time | datetime  |                             |


## redis

* hash key: "provider:`providerType`:resource:`resourceType`"
eg: "provider:aliyun:resource:rds"

field: `id` // eg: r-j6c3a7632edd7094

value:
{
    instance_id: string, // eg: aliyun_r-j6c3a7632edd7094
    id: string, // eg: r-j6c3a7632edd7094
    instance_name: string,
    status: string, //实例运行状态
    type: string，
    provider: string,
}