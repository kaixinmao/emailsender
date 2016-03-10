### 邮件地址格式

```javascript
{
  'addr': 'janz.zhang@toursforfun.cn', //邮件地址
  'name': 'janz.zhang' //别名
}
```


### 发送邮件
> POST /emails

- 输入

| 字段 | 类型 | 描述 | 必填 |
| --- | --- | --- | --- |
| to  | JSON数组  | 发送邮件地址json数组 包含多个地址，最多支持10个，具体参考邮件地址格式| 是 |
| from| 邮件地址| 来源地址| 是 |
| subject | string | 标题 | 是 |
| content | string | 邮件内容 | 是 |
| app_id  | string | 应用来源 | 是 |
| priority | int | 优先级0-100 默认为0 | 否 |
| type | string  | 邮件类型:flash(实时),normal(普通),forad(营销)，默认`normal` | 否 |

- 输出
```javascript
{
  "code":0,
  "msg" : "",
  "data": 1234 //邮件发送的ID号
}
```



### 查询邮件
- 输入
> GET /emails

| 字段     | 类型 | 描述 | 必填 |
| ---      | --- | --- | --- |
| app_id   | string  | 应用来源 | 是 |
| page     | int  | 页数，默认为1，从1开始| 否 |
| per_page | int  | 每页数量，默认20， 最多50| 否 |
| email_id | int  | 要查询的EmailID | 否 |
| start_created   | string | 起始查询创建时间 2016-02-23 默认 endtime 7天前| 否 |
| end_created   | string | 结束查询创建时间 2016-02-23 默认当前时间 | 否 |
| to | string | 收件人地址信息, like %to% 方式查询收件人信息 | 否 |
| search | string | 搜索值，like %search% 方式查询subject| 否 |
| status | int | 状态值，默认为全部, 0待发送，1发送中，10发送成功，-10发送失败| 否 |

- 输出
```javascript
//按时间逆序显示（最新的排在前面）
//{"code":0, "msg":"", "data":具体数据参考下面数据格式}
{
    "total" : 100, //邮件返回总数，超过1000最多只显示1000条
    "emails": {
      "1234" : { //email id 为key
        "subject":"这是邮件标题",
        "to": [{"addr":"janz.zhang@toursforfun.cn", "name":""}],
        "from: {"addr":"service@toursforfun.cn", "name":""},
        "content": "",//邮件内容
        "priority": 15,
        "type" : "forad",
        "status": 0,// 0待发送，1发送中，10发送成功，-10发送失败
        "error" : "发送失败时的错误信息",
        "send_time":12312312//未发送成功为0 时间戳
        "create_time": 12312312 时间戳
      },
      //...其他邮件
    }
}
```

### 马上发送一封邮件
马上发送一个已经存在的邮件ID(不管状态如何都会转为发送中状态)