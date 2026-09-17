# 预编译包升级脚本

install-prebuilt.sh 随 Linux x86_64 预编译包一起分发，应在完整包根目录运行；单独从源码目录运行不适用。包必须包含 gophish、ui-files.txt、migrations.txt 和 SHA256SUMS.txt。

先停止旧服务，再执行：

```bash
bash install-prebuilt.sh /实际原安装目录 --stopped
```

无需 Go/gcc。脚本校验文件、检查架构及原进程、完整备份原安装目录，再替换程序和明确列出的后台页面/脚本，补充缺失迁移。保留配置、数据库和其他文件。外置数据库需单独备份；自行改过被替换的界面文件时需先合并。

程序启动时执行未应用的结构迁移；不包含 SQLite 到 MySQL 的数据迁移。按原服务方式手动重启一个实例。
