<img src="https://user-images.githubusercontent.com/7601383/219330724-55bc04e5-3f19-433b-98fe-6d19333af8b5.png" width="80%" height="80%">  


基于Headscale实现的具有WebUI的Tailscale控制服务器       
    
**注意** 此版本可能与Headscale、Tailscale官方版本均有不兼容的情况，如需与官方版本并用需要考虑进行兼容性测试。    
      
      
# 使用    
~~example.mirage目录是示例配置，请将其重命名为.mirage放置在服务运行同目录下，并将其内acl和config文件名中的example去掉。根据自己情况修改acl和config文件配置。~~    
    
**当前已实现免配启动！！**    
编译后，直接运行程序即可   
* 默认超管控制台监听:8081端口，如需避免冲突进行修改只需运行前配置环境变量，例如 ```export MIRAGE_SYS_ADDR="127.0.0.1:18081"```   
* 建议在程序同目录放置acl.hujson文件便于修改ACL【后续ACL会做进组织数据库中就不需要了】    
* 首次启动因为没有适当配置，控制器服务不会启动，需登录超管控制台（即通过前述的端口）绑定超级管理员（使用WebAuthn方式，建议使用苹果设备或yubikey，现代浏览器的密钥管理系统应该能支持但是并不建议）   
    
     
# 完成功能进度(ToDo List)：    
- [x] 系统零配置文件启动，页面初始化功能   
- [ ] 超管驾驶舱   
    - [x] 超管绑定/登录/登出     
    - [x] 系统运行状态展示与控制    
    - [x] 配置页   
        - [x] 系统基本配置   
        - [x] 三方服务配置【后续需要进行调整】   
        - [x] 超管换绑   
    - [ ] DERP页   
        - [ ] 全局DERP服务配置  
        - [ ] DERP服务自动化部署   
        - [ ] DERP详情页   
    - [ ] 租户页   
        - [ ] 租户列表展示   
        - [ ] 添加/删除租户(与身份鉴别服务配置相关联)   
        - [ ] 租户用量信息   
    - [ ] 用户页   
        - [ ] 个人租户信息展示与管理？  
    - [ ] 系统日志页   
- [x] 注册与登录   
    - [x] Google、Microsoft、GitHub、Apple账号登录
    - [x] ~~OIDC对接阿里云IDaaS登录【暂时弃用】~~   
    - [x] ~~对接阿里云手机号注册【暂时弃用】~~   
    - [ ] 个人微信/QQ/微博登录   
    - [ ] 企业微信/钉钉对接登录    
- [x] 主界面框架与页头部用户基本信息展示、登出     
- [x] 设备页签      
    - [x] 设备列表信息展示      
    - [x] 单个设备详情页展示      
    - [x] 设备IP复制与设备控制菜单展示      
    - [x] 修改设备名      
    - [x] 启/停用设备密钥过期      
    - [x] 编辑设备子网路由    
    - [x] 删除设备      
    - [x] 编辑设备ACL标签   
    - [ ] 分享设备     
    - [ ] 客户端版本更新提示   
- [ ] 服务页签【暂不考虑】    
- [x] 用户页签   
    - [x] 用户列表展示   
    - [x] 管理员身份转移/添加
    - [ ] 更多种类角色划分编辑
    - [ ] 用户设备筛查
    - [ ] 用户冻结
- [ ] ACL页签      
- [x] DNS页签      
    - [x] 基本信息展示      
    - [x] 启/停用MagicDNS     
    - [x] 启/停用Override Local    
    - [x] 添加全球域名服务器   
    - [x] 添加split域名服务器   
    - [x] 修改/删除域名服务器   
    - [x] Basedomain解绑用户名以及可修改   
    - [ ] ~~Split域名服务器可调整顺序【暂不考虑】~~   
    - [ ] ~~DNS厂商预植【暂不考虑】~~   
    - [ ] ~~HTTPS证书（BETA）【暂不考虑】~~    
    - [ ] ~~NextDNS支持【暂不考虑】~~    
- [x] 设置页签       
    - [x] 通用设置 （暂时不处理设备需预授权开关）
        - [x] 组织名称
        - [x] 身份提供商
        - [x] 设备默认秘钥过期时长
    - [x] 标签管理【需要将这个移到ACL页面】      
    - [ ] 特性功能开关    
    - [ ] 账单及用量显示   
    - [ ] 密钥管理   
        - [x] 授权密钥管理【需要考量这个与用户绑定关系】    
        - [ ] ~~API密钥管理【暂不考虑】~~     
    - [ ] ~~webhook【暂不考虑】~~  
    - [ ] ~~OAuth Client【暂不考虑】~~    
- [x] 组织切分及身份源管理   
    - [x] 按组织划分最终用户    
    - [x] 组织管理员配置   
    - [x] DNS、ACL及其他有关系统配置归属至组织   
- [ ] ~~日志页签【暂不考虑】~~    
- [ ] ~~i18n【暂不考虑】~~   
- [ ] ~~前后端分离【是否有必要？】【暂不考虑】~~      
    
  
# 截图    
    
## 超管界面    
    
<img src="https://user-images.githubusercontent.com/7601383/226161921-0df684fa-4956-4681-b5d1-119e5f434246.png" width="50%" height="50%">
    
    
## 管理员（当前是用户）界面    
       
<img src="https://user-images.githubusercontent.com/7601383/218957899-89ba2492-8508-40aa-b6c7-92d57ecde6d5.png" width="80%" height="80%">   
<img src="https://user-images.githubusercontent.com/7601383/218959833-b2e70903-5250-4fd1-b175-a6b29aef9199.png" width="80%" height="80%">    
<img src="https://user-images.githubusercontent.com/7601383/218959838-2fc1ba9d-b372-4890-806e-28a0ce5c4928.png" width="80%" height="80%">    
<img src="https://user-images.githubusercontent.com/7601383/218959854-b3d02123-b917-4f06-a8cb-a619f591cd42.png" width="80%" height="80%">    
<img src="https://user-images.githubusercontent.com/7601383/218959858-83f4f57e-dfa6-4886-bd1d-e608b8e7b27c.png" width="80%" height="80%">    
<img src="https://user-images.githubusercontent.com/7601383/218959865-35cc93ea-953a-451e-881c-8b216518971d.png" width="80%" height="80%">    
<img src="https://user-images.githubusercontent.com/7601383/218959876-32b0c444-8be4-4372-89f8-05a50c9f775c.png" width="80%" height="80%">    
<img src="https://user-images.githubusercontent.com/7601383/226161978-04a80f9a-5a79-4a09-823a-e76d921cd629.png" width="50%" height="50%">
<img src="https://user-images.githubusercontent.com/7601383/226162115-2aa9aad3-48aa-4a86-8608-4d82058f07fa.png" width="50%" height="50%">
<img src="https://user-images.githubusercontent.com/7601383/226162138-7a5f2d4b-f699-40da-a2a2-7cd7f9dd8c39.png" width="50%" height="50%">
<img src="https://user-images.githubusercontent.com/7601383/218959888-36064015-1052-454b-a5ad-854ceaabe363.png" width="80%" height="80%">    
<img src="https://user-images.githubusercontent.com/7601383/218959893-8c398c91-bddc-4176-836f-ebc6806567b0.png" width="80%" height="80%">    
    

      
感谢Tailscale、Headscale及开源社区的共同努力。    
      
