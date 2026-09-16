# Gophish 中文增强版：长期演练

基于 Gophish 0.12.1 的授权安全演练项目，保留本地中文界面、二维码与投递通道相关改动，新增长期演练和多用户组自动同步。

## 长期演练

- 一个演练可关联多个用户组，组内新增人员由后台 worker 约每分钟自动加入发送队列。
- 同一个组可以关联多个独立演练，各演练使用自己的模板、落地页和投递配置。
- 单个演练内按邮箱去重，忽略大小写与首尾空格；已有链接、发送记录和历史结果保持不变。
- 支持已有未完成演练开启长期模式，以及手动追加人员和指定本次追加的发送时间。
- 关联使用组 ID，改名继续生效；解绑或关闭长期模式停止后续自动加入，已排队的投递保留。

详细操作、接口与行为边界见 [长期演练使用说明](LONG_TERM_CAMPAIGNS_CN.md)。

## 构建与测试

需要 Go 和 C 编译器（SQLite 驱动依赖 CGO）。仓库包含 vendor 依赖，可以离线构建：

```bash
CGO_ENABLED=1 GOPROXY=off GOSUMDB=off go build -mod=vendor -o gophish .
CGO_ENABLED=1 GOPROXY=off GOSUMDB=off go test -mod=vendor ./models ./controllers/api ./worker
```

首次运行：

```bash
cp config.example.json config.json
# 按部署环境编辑 config.json
./gophish --config config.json
```

示例管理端口为 3333，演练入口端口为 8081。运行时配置、数据库、日志和本地编译产物不纳入 Git。

服务器升级与数据保留说明见 [部署说明](SERVER_DEPLOY_CN.md)。

## 验证范围

交接版本已完成 Go 构建、SQLite 上 models/controllers/api/worker 回归测试和前端脚本语法检查。尚未验证真实 MySQL、浏览器端到端流程或真实邮件/钉钉投递。本仓库同步不代表服务器已部署。

## 版本来源

初始源码来自 2026-09-16 长期演练更新包，源码内容与本地 `gophish-0.12.1` 对应文件一致。更新包 SHA256：

```text
11a23c460dabfff6402100eb4f190c6abfaa7a7a82d1fa70853d77d5ab1211d5
```

上游说明见 [Gophish README](docs/README_UPSTREAM.md)，许可证见 [LICENSE](LICENSE)。
