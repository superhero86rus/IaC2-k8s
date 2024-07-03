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
```

### Развертывание через kubeadm
```bash
# Генерируем ключ на node1 и копируем его на другие ноды
ssh-keygen
ssh-copy-id node2
ssh-copy-id node3
```

```bash
# Ставим docker на всех узлах (через bash)
bash -c '
apt install docker.io -y
ssh node2 apt install docker.io -y
ssh node3 apt install docker.io -y
'

# Ставим GitLab runner (не в контейнере) на сервер
wget http://gate.isp.un/unix/Git/gitlab-runner_amd64.deb
dpkg -i gitlab-runner_amd64.deb

# Подключаем виртуальный адаптер хоста к клиентской машине client1
# Выполняем команду
dhclient eth0

# Кубернетес может работать только на узлах, на которых отключен swap
bash -c '
swapoff -a
ssh node2 swapoff -a
ssh node3 swapoff -a
'

# Делаем так, чтобы swap не включился после перезагрузки
bash -c '
sed -i"" -e "/swap/s/^/#/" /etc/fstab
ssh node2 sed -i"" -e "/swap/s/^/#/" /etc/fstab
ssh node3 sed -i"" -e "/swap/s/^/#/" /etc/fstab
'

bash -c '
http_proxy=http://proxy.isp.un:3128 apt install apt-transport-https curl -y
ssh node2 http_proxy=http://proxy.isp.un:3128 apt install apt-transport-https curl -y
ssh node3 http_proxy=http://proxy.isp.un:3128 apt install apt-transport-https curl -y
'

# Устанавливаем пакеты k8s (выполнить на каждой ноде)
bash -c '
mkdir /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.28/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.28/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt update
sudo apt install -y kubeadm=1.28.1-1.1 kubelet=1.28.1-1.1 kubectl=1.28.1-1.1
'

# Узнаем какой ip у client1 через gate
dhcp-lease-list
#MAC                IP              hostname       valid until         manufacturer
#===============================================================================================
#08:00:27:66:1e:dc  192.168.18.101  debian         2024-07-03 11:57:33 -NA-

# Пробрасываем ключ с сервера на client1
ssh-copy-id 192.168.18.101
# Настраиваем клиент
ssh 192.168.18.101
nano /etc/hostname
# Прописываем там имя client1

nano /etc/hosts
# Добавляем
127.0.1.1       client1

nano /etc/network/interfaces
# Раскомментировать строчку #iface eth0 inet static

init 6
```

#### Инициализация master (node1)
```bash
# Добавить кластерную сеть (node1)
kubeadm init --pod-network-cidr=10.244.0.0/16 --apiserver-advertise-address=192.168.18.201

# Копируем файл конфигурации (созданный kubeadm) в каталог настроек
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config

# Настраиваем кластерную сеть
kubectl apply -f https://raw.githubusercontent.com/coreos/flannel/master/Documentation/kube-flannel.yml

# Посмотреть список подов
kubectl get pod -o wide --all-namespaces

# Статус кластера
kubectl get --raw='/readyz?verbose'
```

#### Подключение рабочих узлов (node2, node3)
```bash
# Теперь мы должны выполнить команду, которую получили после создания кластера, на воркер нодах
kubeadm join 192.168.18.201:6443 --token 64z3ah.gblgfg1z3qgcmaee \
        --discovery-token-ca-cert-hash sha256:1ab01b5f5f187677b8c3610caed049345e9a1aa3d9edf0d87b3bf796b178de8d
```

#### Добавляем ноды второго кластера, разворачиваемого через Kubespray
```bash
# Делаем на GATE все что ниже

nano /etc/bind/corp18.un
# Добавить $GENERATE 1-9 kube$ A 192.168.18.22$
service named restart
```

```bash
nano /etc/dhcp/dhcpd.conf

# Вставляем (предварительно поменяв мак-адреса у 4-х машин)
host kube1 {hardware ethernet 08:00:27:12:34:51;fixed-address 192.168.18.221;}
host kube2 {hardware ethernet 08:00:27:12:34:52;fixed-address 192.168.18.222;}
host kube3 {hardware ethernet 08:00:27:12:34:53;fixed-address 192.168.18.223;}
host kube4 {hardware ethernet 08:00:27:12:34:54;fixed-address 192.168.18.224;}

service isc-dhcp-server restart

# Подключаемся к 4м нодам и получаем адреса
dhclient eth0
```

```bash
# Настраиваем кластер через ansible
# Создаем на сервере конфигурацию для ансибла
mkdir /etc/ansible
touch /etc/ansible/hosts
```

```ini
[all:vars]
ansible_python_interpreter="/usr/bin/python3"

[kubes]
kube[1:4]

[kubes:vars]
ansible_ssh_user=root
ansible_ssh_pass=Pa$$w0rd
# Повышение привилегий не нужно
# ansible_become=yes
```

```bash
# Установка sshpass (для захода без ключа, а только по паролю)
apt install sshpass

# Отключаем проверку валидности ключа
nano /etc/ansible/ansible.cfg
# [defaults]
# host_key_checking = False

# Проверка связи
ansible all -m ping

# Отключаем swap на узлах
ansible kubes -a 'sed -i"" -e "/swap/s/^/#/" /etc/fstab'
ansible kubes -a 'swapoff -a'

# Настраиваем ansible
cd conf/ansible/roles
nano nodes.yml
# Меняем группу на kubes

cd node/vars
nano main.yml
# Меняем префикс на kube
# Добавляем -1 чтобы номер брать из последней цифры ip-адреса
```
```yml
---
# vars file for node
name_prefix: kube
X: "{{ ansible_eth0.ipv4.address.split('.')[2] }}"
N: "{{ ansible_eth0.ipv4.address.split('.')[3][-1] }}"
```
```bash
# Оставляем один dns сервер
nano /node/templates/resolv.conf.j2
# nameserver 192.168.{{ X }}.10

nano /node/templates/interfaces.j2
# gateway 192.168.{{ X }}.1

# Теперь можно выполнить playbook и сконфигурировать сеть
ansible-playbook -f 5 conf/ansible/roles/nodes.yml
```
```bash
# Генерируем ключ на kube1 и переносим его на ноды
ssh-keygen
ssh-copy-id kube2
ssh-copy-id kube3
ssh-copy-id kube4
```