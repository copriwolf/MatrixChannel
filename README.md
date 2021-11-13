# 🕹 MatrixChannel
📮  Sync Tapd To Notion. 同步 Tapd 数据至 Notion。

`中文版`




## 🥞 特性

✏️  简单配置：填充配置文件即可同步数据

🗃 多模块同步：可同步 Tapd 需求模块数据、任务模块数据

👨‍👨‍👦‍👦 多用户支持：支持一次任务同一个企业下的多用户同步

🔮 OAuth & Callback 支持：支持官方操作多账户的公共 Bot，提供 OAuth 验证与回调页面



## 🧐 先决条件

1. 具备 Tapd Api 账号、密码 [[Tapd官方配置说明](https://www.tapd.cn/help/show#1120003271001000093)]
2. 生成了私人、公共 Notion Bot 的密钥 [[Notion API 快速入门说明](https://developers.notion.com/docs/getting-started)]


## 🏄‍♂️ 快速上手

1. 点击右方链接，拷贝模版到你自己的 Notion 中。 [[📍 Tapd 模版](https://www.notion.so/copriwolf/Tapd-f85af3ec57154292be411242e8a33122)]
    ![copyTmpl](https://user-images.githubusercontent.com/10501324/141659707-e3c49a5b-5c04-4fd2-b6e3-e35eea9859bb.png)


2. 授权你的机器人访问与修改该页面数据。

    ![Kapture 2021-11-14 at 05 39 21](https://user-images.githubusercontent.com/10501324/141659781-25da7d2c-c216-44a7-b40f-0221326474fe.gif)


3. 获取页面的 PageID
    > 由于 Notion Api 处于 Beta 版本，未开放处理 Database 级别权限，所以需要你手动复制 DataBase 的 ID。
    
    ![screenshot](https://user-images.githubusercontent.com/10501324/141659692-34e3a5a8-0dd7-4898-97ee-23207e11059f.gif)
    分别进去【TapdStory需求表】与【TapdTask任务表】页面中，在右上角的 【Share】-【Copy Link】，在第两个 / 与 ?v=之间的即为 DataBaseID。

    ```bash
    https://www.notion.so/copriwolf/84999c421caf4eeeab8bc66bc044408a?v=9...
                                    <---------- DataBaseID --------->
    ```                                

4. 复制项目中的 `config/demo.conf.yaml` 为 `conf.yaml`，并在其中填充自定义的数据。

   ```yaml
   Service:
       # 数据同步间隔
       refreshInterval: 10m0s
       # 请求失败的睡眠时间
       httpRequestFailSleepTime: 5s
       # 最大请求失败重试次数
       httpRequestAttempts: 5
       # Tapd Api 用户名
       tapdApiUser: vvvvvv
       # Tapd Api 密码
       tapdApiPassword: B8888888-8888-9999-0000-SSSSSSSC
       # Tapd 公司 ID
       tapdCompanyID: "0000700"
   
   User:
       # 用户 A 的昵称/电邮（任意..）
       copriwolf:
           enable:
               - task
           # 该用户在 Tapd 中的用户名
           tapdOwner: copriwolf
           # Notion 访问密钥（私有机器人使用 Bot Secret，多人共享机器人使用用户授权 AccessToken）
           notionBotSecret: secret_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
           # Notion Database 任务页面的 ID
           notionDbTaskID: 77777777777777777777777777777777
   
   ```

   

5. 执行命令，以 docker 形式运行

   ```bash
   $ docker run -it \
   -v "$PWD/conf.yaml:/app/config/conf.yaml" 
   ghcr.io/copriwolf/matrixchannel:master
   ```

   


