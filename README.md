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

```bash
vagrant box add /c/distrs/ubuntu_20.04.box --name specialist/ubuntu20
# В клонированном репозитории нужно прописать стенд
cd ~/conf/vagrant/nodes
nano Vagrantfile
# Поднимаем машины nodeX
time vagrant up --parallel
```

```bash
# Настраеваем прокси через gate для server (вбиваем на сервере)
export http_proxy=http://proxy.isp.un:3128
export https_proxy=http://proxy.isp.un:3128
export no_proxy=localhost,127.0.0.1,isp.un,corpX.un # Не проксировать (исключение)
```

```bash
# Ставим docker на сервер
apt install docker.io
apt install docker-compose

# Ставим ansible на сервер
apt install ansible

# Ставим GitLab runner на gate (запускаем через контейнер docker)
docker run -d --name gitlab-runner --restart always \
  -v /srv/gitlab-runner/config:/etc/gitlab-runner \
  -v /var/run/docker.sock:/var/run/docker.sock \
  gitlab/gitlab-runner:latest

# Настраиваем машину client1_03.07.2024 импортируя образ C:\Distrs\debian_11.1_64_01.ova в VirtualBox
```

```bash
# Установливаем GitLab через docker-compose.yml на сервер
version: '3.6'
services:
  web:
    image: 'gitlab/gitlab-ce:latest'
#    image: 'gitlab/gitlab-ce:16.7.4-ce.0'
    restart: always
    hostname: 'server.corp18.un'
    environment:
      GITLAB_ROOT_PASSWORD: "strongpassword"
      GITLAB_OMNIBUS_CONFIG: |
        # Можно задать пароль так, вместо GITLAB_ROOT_PASSWORD
        # gitlab_rails['initial_root_password'] = 'strongpassword'
        prometheus_monitoring['enable'] = false
        gitlab_rails['registry_enabled'] = true
        gitlab_rails['registry_host'] = "server.corp18.un"
        external_url 'http://server.corp18.un'
        registry_external_url 'http://server.corp18.un'
        gitlab_rails['registry_port'] = "5000"
        registry['registry_http_addr'] = "server.corp18.un:5000"
#        external_url 'https://server.corpX.un'
#        registry_external_url 'https://server.corpX.un:5000'
#        gitlab_rails['registry_port'] = "5050"
#        registry['registry_http_addr'] = "server.corpX.un:5050"
    ports:
      - '80:80'
#      - '443:443'
      - '2222:22'
      - '5000:5000'
    volumes:
      - '/etc/gitlab:/etc/gitlab'
      - '/srv/gitlab/logs:/var/log/gitlab'
      - '/srv/gitlab/data:/var/opt/gitlab'
    shm_size: '256m'
    logging:
      driver: "json-file"
      options:
        max-size: "2048m"
```

```bash
# Добавить имена узлов на server
nano /etc/bind/corp8.un
# Добавляем узлы (до 9 штук) после строки начинающейся на server
$GENERATE 1-9 node$ A 192.168.18.20$

# Перезапуск dhcp
service named restart
# Проверяем
nslookup node1
ping node1

# Добавляем ключи сервера на ноды
ssh-copy-id node1
ssh-copy-id node2
ssh-copy-id node3
```
