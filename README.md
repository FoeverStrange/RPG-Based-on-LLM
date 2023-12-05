# RPG-Based-on-LLM——Fantastic Life奇妙人生
Fantastic Life奇妙人生是一个基于大语言模型的RPG游戏，在此游戏中你可以扮演您想要成为的任何人，在您想要的世界观中走过一段故事


## 玩法介绍

### RPG游戏
在聊天对话框中，选择你想要游玩的故事设定，并通过文字对话的形式做出行动，影响剧情走向。

可供选择的故事有：

| 名称     | 解释 | prompt                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| ---------- | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 哈利波特 |      | 请你为我提供一个文字类型的RPG游戏，游戏过程中会根据我的选择推进故事的发展，<br />你将作为这个故事的讲述者，并为我生成描绘故事以供我选择我要做什么，我将会说出我的决策，以供你生成下一段故事<br />我希望以《哈利波特》的世界观作为故事背景，我想要扮演一个和哈利波特同岁的魔法师，与哈利波特、马尔福同一年入学，根据我的选择的不同我可能会成为哈利波特的重要伙伴，或是黑魔法师中的一员，甚至代替哈利波特成为救世主。<br />一次只描绘一段故事，每当你描绘完一段故事后，请你停止生成，并等待我的选择 |
| 斗罗大陆 |      |                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |

### 塔罗占卜
使用塔罗牌进行占卜

| 牌型   | 解释 | prompt                                                                                        |
|------| ------ |-----------------------------------------------------------------------------------------------|
| 三牌   |      | 请你为我随机生成1~22中的三个不重复的、随机排序的数字，每个数字代表塔罗牌中的一张牌，请告诉我你生成的数字并对应塔罗牌中的牌；继续： 这三张牌预示着我的过去、现在和将来，请你为我解读这三张牌的寓意 |
| 命运十字 |      |                                                                                               |

### 聊天建议
作为恋爱专家，我希望你能够为我提供一些聊天建议，帮助我和我的暧昧对象更好的沟通


| 名称 | 解释 | prompt                                                                                        |
|----| ------ |-----------------------------------------------------------------------------------------------|
| 恋爱专家 |      | 请你担任我和暧昧对象聊天的建议官，我会给你发送我的暧昧对象说了什么，请你帮助我分析对方说这句话时候的态度，以及潜意识中可能想表达的信息，并为我生成回复对方的建议 |

## 部署方法

### config
/config/config_template.yaml 修改为 /config/config.yaml，并填写其中的配置信息



## TODO List

* [X] 2023年12月2日：搭建基础框架Gin
* [X] 2023年12月2日：使用api调试，实现和百川模型的对话功能，完成单次发送消息的函数
  * [X] 2023年12月4日：根据c *Gin.context发送单次消息
  * [X] 2023年12月5日：完成多轮对话功能
* [X] 2023年12月3日：logger更新为zap.logger
* [X] 2023年12月5日：架构设计，模块化对话功能平面，维护多个Bot对象，每个Bot对象维护一个对话，并且有初始化对话和继续对话的功能
  * [X] 2023年12月3日：抽象出对话功能接口、用户管理接口；会话功能平面待完善
  * [X] 2023年12月4日：会话相关模块架构设计
  * [X] 2023年12月5日：完善会话管理
* [ ] 缝合IM系统
  * [ ] 用户管理模块
  * [ ] 多服务器gRPC协同
* [ ] docker化
* [ ] 将模型部署在云服务器上验证基本对话功能
* [ ] prompt模板嵌入
* [ ] GPT支持

