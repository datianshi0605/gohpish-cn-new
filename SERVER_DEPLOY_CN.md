# 服务器构建与升级

本仓库提供源码和 vendor 依赖，不包含预编译程序、实际配置或数据库。

## 新环境

1. 安装 Go 和 gcc，在仓库根目录运行 `bash build-linux-on-server.sh`。
2. 将 `config.example.json` 复制为 `config.json`，根据部署环境设置监听地址、TLS 和数据库。
3. 在仓库根目录运行 `./gophish --config config.json`。管理端默认端口为 3333，演练入口默认端口为 8081。管理端应按环境设置访问控制。

## 已有环境升级

先确认服务器现有版本、安装目录、启动方式、数据库位置及演示内容目录，再停止服务并完整备份。外置数据库/MySQL 需单独备份。不要用示例配置或空数据库替换原数据。

长期演练升级涉及新编译的程序及以下界面文件；已有界面定制需先比对合并：

- `templates/campaigns.html`
- `templates/campaign_results.html`
- `static/js/src/app/campaigns.js`
- `static/js/dist/app/campaigns.min.js`
- `static/js/src/app/campaign_longterm.js`

还需补齐 `db/db_sqlite3/migrations` 或 `db/db_mysql/migrations` 中缺失的迁移，已有迁移如不一致需先核对，不能直接覆盖。长期演练新增迁移为 `20260916000000_long_term_campaign.sql` 和 `20260916010000_campaign_auto_groups.sql`。启动时原有迁移机制将新增字段和关联表。

保留原配置、数据库和演示文件，使用原启动方式重启。已有活动默认不自动开启长期模式；先按 [使用说明](LONG_TERM_CAMPAIGNS_CN.md) 用测试组验证。

当前未执行服务器部署，未在真实 MySQL 实例验证。
