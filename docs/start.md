# 部署

### p2p引导节点的部署
    
    1.编译cmd/tools
    2.在服务器上启动该程序，
        -k 指定私钥文件
        -p 私钥文件密码
        
        例：
        ./tools -k ../wallet/keystore/UBkZ43E3rwC2mYKfGfUHH9qnQdC7hpkbA5H.json -p 1
       
    3.如果需要链接该引导节点，可以在配置文件的boot项中配置该引导节点的地址信息
       
        例：
        Bootstrap="/ip4/111.111.111.111/tcp/2211/ipfs/16Uiu2HAky7wHeSacNDy9sooSktQ2hjSULJEfqzrU47k8nnr1JR5v"
      
### 超级节点的部署

    1.部署超级节点需要在配置文件中配置超级节点的地址的私钥文件和私钥文件的解密密码
       
        例：
        # If it is a block generating node, it needs to be configured
        # Json file address of the address private key
        KeyFile = "../wallet/keystore/UBkZ43E3rwC2mYKfGfUHH9qnQdC7hpkbA5H.json"
        # Password to decrypt the private key json file
        KeyPass = "1"
    2.配置节点的外部ip（必须配置）
        
        例：
        # P2P configuration
        #Externale ip (default = "0.0.0.0")
        ExternalIp = "1.2.3.4"
        
    3.配置rpc密码
    
        例：
        # RPC password
        RpcPass = "123456"
    4.配置节点数据存储目录
    
        例：
        HomeDir = "data"
    5.使用配置文件启动
        ./UWORLD --config config.toml
   
      