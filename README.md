# GDUT网络测试工具

- 网站：https://network.gdutnic.com/ （仅限校园网访问）  
- 该项目由网管队技术组开发，校园网用户可以运行测试工具对当前校园网各出口进行延迟和丢包率测试。所有测试结果会上传到网站上（https://network.gdutnic.com/），服务器进行统计后显示出来。用户可以根据这些数据**粗略**判断校园网出口的状况。（我们正努力向网络中心争取各出口带宽使用率的权限，目前暂时无法向大家提供各出口的准确数据）  
- 运行程序后会提供两个选项：单次测试和持续测试。我们鼓励用户使用持续测试，为大家持续提供可靠的数据。  
- 本测试工具采用GPLv3开源协议。服务端考虑到安全问题，为了防止恶意构造的数据包攻击，暂不开源。  

# 编译

注：编译过程中如发现有库缺失，请自行`go get`安装相关库。

```shell
cd $GOPATH/src
git clone https://github.com/gdut-network-manager/netter.git --depth=1
cd netter
go build
```
