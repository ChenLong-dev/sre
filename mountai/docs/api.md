### 获取项目的资源 API

#### 1.1 获取项目的资源

```
GET /v1/projects/:projectID/resources
```

Request:

| params | type   | required | desc |
| ------ | ------ | -------- | ---- |
| env    | string | true     |      |

```
{
    errcode: int,
    errmsg: string,
    data: {
        _id: string,
        team_id: string,
        project_id: string,
        rds: [{
            instance_id: string,
            instance_name: string,
            status: string, //实例运行状态
            type: string，
            provider: string,
        },],
        redis: [{
            instance_id: string,
            instance_name: string,
            status: string, //实例运行状态
            type: string,
            provider: string,
        },],
        mongo: [{
            instance_id: string,
            instance_name: string,
            status: string, //实例运行状态
            type: string，
            provider: string,
        },],
        hbase: [{
            instance_id: string,
            instance_name: string,
            status: string, //实例运行状态
            type: string，
            provider: string,
        },],
        cdn: [{
            instance_id: string,
            instance_name: string,
            status: string, //实例运行状态
            type: string，
            provider: string,
        },],
        commit_id: string,
        env: string,
        updated_time: string,
        created_time: string,
    },
}
```

#### 1.2 修改项目的资源

```
PUT /v1/projects/:projectID/resources/:resourceID
```

Request:

| params | type     | required | desc        |
| ------ | -------- | -------- | ----------- |
| rds    | [string] | false    | instance_id |
| redis  | [string] | false    | instance_id |
| mongo  | [string] | false    | instance_id |
| hbase  | [string] | false    | instance_id |
| cdn    | [string] | false    | instance_id |

```
{
    errcode: int,
    errmsg: string,
    data: {
        _id: string,
        team_id: string,
        project_id: string,
        rds: [string],
        redis: [string],
        mongo: [string],
        hbase: [string],
        cdn: [string],
        commit_id: string,
        env: string,
        updated_time: string,
        created_time: string,
    },
}
```

#### 1.3 获取资源实例列表

```
GET /v1/resources
```

| params   | type   | required | desc                                      |
| -------- | ------ | -------- | ----------------------------------------- |
| type     | string | true     | `rds` `mongo` `ecs` `hbase` `cdn` `redis` |
| provider | string | false    | `aliyun`                                  |

```
{
    errcode: int,
    errmsg: string,
    data: {
        total: int,
        items: [{
            instance_id: string,
            id: string,
            instance_name: string,
            status: string, //实例运行状态
            type: string，
            provider: string,
        },],
    },
}
```

### 收藏项目 API

#### 2.1 创建收藏项目

```
POST /v1/fav_projects
```

Request:

| params     | type   | required | desc |
| ---------- | ------ | -------- | ---- |
| project_id | string | true     |      |

```
{
    errcode: int,
    errmsg: string,
    data: {
        id: string,
        user: {
            id: string,
            name: string,
            avatar_url: string,
            email: string,
        },
        project: {
            id: string,
            name: string,
            language: string,
            desc: string,
            api_doc_url: string,
            dev_doc_url: string,
            labels: [string],
            team: {
                id: string,
                name: string,
            },
        },
    }
}
```

#### 2.2 删除收藏项目

```
DELETE /v1/fav_projects/:id
```

```
{
    errcode: int,
    errmsg: string,
    data: null,
}
```

#### 2.3 获取收藏项目列表

```
GET /v1/fav_projects
```

Request:

| params     | type   | required | desc                                  |
| ---------- | ------ | -------- | ------------------------------------- |
| project_id | string | false    |                                       |
| user_id    | string | false    | 不传时，获取的当前 session 用户的收藏 |

```
{
    errcode: int,
    errmsg: string,
    data: {
        total: int,
        items: [{
            id: string,
            user: {
                id: string,
                name: string,
                avatar_url: string,
                email: string,
            },
            project: {
                id: string,
                name: string,
                language: string,
                desc: string,
                api_doc_url: string,
                dev_doc_url: string,
                labels: [string],
                team: {
                    id: string,
                    name: string,
                },
            },
        }],
    }
}
```

#### 2.4（改动）获取项目动态

```
GET /api/v1/activities
```

增加了支持查询收藏项目动态的字段

| param  | type | desc                       |
| ------ | ---- | -------------------------- |
| is_fav | bool | 是否是收藏项目，默认 false |

#### 2.5（改动）获取项目列表

```
GET /api/v1/projects
```

返回的 list 中增加了是否是收藏项目的字段

| param  | type | desc                       |
| ------ | ---- | -------------------------- |
| is_fav | bool | 是否是收藏项目，默认 false |

#### 2.6（改动）获取项目详情

```
GET /api/v1/projects/:id
```

返回的详情中增加了是否是收藏项目的字段

| param  | type | desc                       |
| ------ | ---- | -------------------------- |
| is_fav | bool | 是否是收藏项目，默认 false |
