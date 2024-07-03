# IaC2-k8s
Курс "DevOps. Уровень 2. Использование Kubernetes". УЦ "Специалист" 03-05.07.2024

### k8s день 1

```txt
Номер стенда NN = 18
https://root:03.07.2024@val.bmstu.ru/video/NNN
RDP: 80.250.209.226:239NN
l: administrator
p: Pa$$w0rd#10
```
```bash
# Методичка
# https://wikival.bmstu.ru/doku.php?id=devops2._%D0%B8%D1%81%D0%BF%D0%BE%D0%BB%D1%8C%D0%B7%D0%BE%D0%B2%D0%B0%D0%BD%D0%B8%D0%B5_kubernetes
git clone http://val.bmstu.ru/unix/conf.git
cd conf/virtualbox
./setup.sh 18 9

# root/Pa$$w0rd
# Заходим на gate 192.168.18.1
sh net_gate.sh
itit 6

sh conf/dhcp.sh

# Заходим на server 192.168.18.10
sh net_server.sh
init 6

sh conf/dns.sh
nano /etc/resolv.conf
# Указываем имя нашего сервера
# nameserver 192.168.18.10

ssh-keygen
ssh-copy-id gate
scp /etc/resolv.conf gate:/etc/
```


