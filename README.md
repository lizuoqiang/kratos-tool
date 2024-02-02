# Install

```
go install github.com/lizuoqiang/kratos-tool/cmd/gen-code@latest
```

# Usage

```
cat /tmp/sql.txt
CREATE TABLE `whitelist` (
    `id` int(10) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键id',
    `city_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '城市id',
    `name` varchar(100) NOT NULL DEFAULT '' COMMENT '名称',
    `start_time` datetime DEFAULT NULL COMMENT '开始时间',
    `end_time` datetime DEFAULT NULL COMMENT '结束时间',
    `customer_ids` text COMMENT '逗号分隔',
    `is_deleted` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '删除状态：0-未删除，1-已删除',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) COMMENT='白名单表';

gen-code /tmp/sql.txt data,model,biz,protobuf,service

output:
生成文件： /tmp/biz/whitelist.go
生成文件： /tmp/model/whitelist.go
生成文件： /tmp/data/whitelist.go
生成文件： /tmp/protobuf/whitelist.proto
生成文件： /tmp/service/whitelist.go

```