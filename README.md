# IaC2-k8s
Курс "DevOps. Уровень 2. Использование Kubernetes". УЦ "Специалист" 03-05.07.2024
val@bmstu.ru

### Полезная статья преподавателя - самый простой пример CI/CI
https://habr.com/ru/articles/716454/

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
### Развертывание через Kubespray
```bash
apt update
apt install python3-pip -y
git clone https://github.com/kubernetes-sigs/kubespray
cd kubespray
time pip3 install -r requirements.txt
cp -rfp inventory/sample inventory/mycluster

# Создание инвентарного файла
pip install ruamel_yaml
declare -a IPS=(kube1,192.168.18.221 kube2,192.168.18.222 kube3,192.168.18.223)
CONFIG_FILE=inventory/mycluster/hosts.yaml python3 contrib/inventory_builder/inventory.py ${IPS[@]}

time ansible-playbook -i inventory/mycluster/hosts.yaml cluster.yml
# После разворачивани кластера, проверяем ноды
kubectl get nodes -o wide
```

### k8s день 2

```bash
# Изучаекм объекты и команды k8s
kubectl get all
# Кубернетес автоматически останавливает образы, в которых отсутствуют процессы
kubectl run my-debian --image=debian -- "sleep" "3600"
kubectl get all

# Подключаемся к поду
kubectl exec -it my-debian exec -- bash
> apt update
> apt install iproute2
> ip a

# Удаление пода
kubectl delete pod my-debian

# Создание deployment (ReplicaSet просят больше не создавать руками, а делать сразу deployment, который создаст сам replicaset)
kubectl create deployment my-debian --image=debian -- "sleep" "3600"
```

### Весь k8s задекларирован с помощью манифестов и мы можем это посмотреть
```bash
kubectl get deployment my-debian -o yaml

# Установим nano как редактор по умолчанию
# Нужно раскомментировать строку в .bashrc
#export EDITOR=nano
kubectl edit deployment my-debian
# Можно увеличить кол-во реплик и изменения прмиенятся моментально

kubectl delete deployment my-debian

# Создаем манифест репликасета
nano my-debian-deployment.yaml
```
```yaml
apiVersion: apps/v1
kind: ReplicaSet
#kind: Deployment
metadata:
  name: my-debian
spec:
  selector:
    matchLabels:
      app: my-debian
  replicas: 2
  template:
    metadata:
      labels:
        app: my-debian
    spec:
      containers:
      - name: my-debian
        image: debian
        command: ["/bin/sh"]
        args: ["-c", "while true; do echo hello; sleep 3;done"]
      restartPolicy: Always
```

```bash
# Применяем манифест деплоймента для репликасета
kubectl apply -f my-debian-deployment.yaml
# Удаляем
kubectl delete -f my-debian-deployment.yaml
```

#### Метки
```bash
kubectl get nodes --show-labels
kubectl label nodes node2 disk2=yes
```

### Volumes
```bash
ssh node2 mkdir /disk2
ssh node2 touch /disk2/disk2_node2
nano my-debian-deployment.yaml

# Добавляем
volumeMounts:
          - name: my-disk2-volume
            mountPath: /data

volumes:
        - name: my-disk2-volume
          hostPath:
            path: /disk2/
      nodeSelector:
        disk2: "yes"

kubectl apply -f my-debian-deployment.yaml

nano my-ha-pv.yaml
# Добавляем Persisten Volume
apiVersion: v1
kind: PersistentVolume
metadata:
  name: my-pv-node2-sz-128m-num-001
  labels:
    type: local
spec:
  storageClassName: my-ha-sc
  capacity:
    storage: 128Mi
  accessModes:
    - ReadWriteMany
  hostPath:
    path: /disk2
  persistentVolumeReclaimPolicy: Retain
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - node2
---
apiVersion: v1
kind: PersistentVolume
metadata:
  name: my-pv-node2-sz-64m-num-002
  labels:
    type: local
spec:
  storageClassName: my-ha-sc
  capacity:
    storage: 64Mi
  accessModes:
    - ReadWriteMany
  hostPath:
    path: /disk2
  persistentVolumeReclaimPolicy: Retain
  nodeAffinity:
    required:
      nodeSelectorTerms:
      - matchExpressions:
        - key: kubernetes.io/hostname
          operator: In
          values:
          - node2

# В отличие от простых томов, к томам Persistent Volume поды не подключаются автоматически
# Поды подключаются через запрос PV Claim
kubectl apply -f my-ha-pv.yaml

nano my-ha-pvc.yaml

apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-ha-pvc-sz64m
spec:
  storageClassName: my-ha-sc
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 64Mi

kubectl apply -f my-ha-pvc.yaml
kubectl get persistentvolume

# Создадим файл сиинтетически, размером больше чем том
kubectl exec -it pods/my-debian-lskkw -- bash
dd if=/dev/zero of=/data/filedisk bs=1M count=100
ls -lh /data/
```

### Как подключиться по SSH к поду k8s
```bash

# Создаем конфиг
nano sshd_config

PermitRootLogin yes
PasswordAuthentication no
ChallengeResponseAuthentication no
UsePAM no

kubectl create configmap ssh-config --from-file=sshd_config --dry-run=client -o yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: ssh-config
data:
  sshd_config: |
    PermitRootLogin yes
    PasswordAuthentication no
    ChallengeResponseAuthentication no
    UsePAM no
  authorized_keys: |
    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCkaacIUANGOcJoXXoX034GuY6Ci1E/iXZnRgR6Lu2SnbZGiGSBfucBONYYScvtwb2cpLkQHXc5ouA1Z25nDbj3r4+AP/LQCbYlvDo96vor29rXu/5SmoacRs+oumIhrQUGv/cQRGDqZzNiK05jnMVySAcVt+Bgw9xwl8SDg8wqGqMfJwl7ZEqEv7lHr93f+EBjSoisTjmOLg0M2AatKKcbN2kXUByIi+TfF4F097zXnrfEjyPsC4yMYXyag4B7McaQgL2Xy2/jUZMZe/tn1RY7AIPoiONN3/qVE64gYlaIouWJEHlycIhZQCxmk95JA1crg9KuD5E8YWpvTlXimj9ydQdNVa5G4XqTQ/sA+kWqCp8EPwKwVxXLKW/1vgYX5U8QmsuNpwYWi7dkifS4SeYmM6ilLvoNsyyvIgsN2EHAHGJTflR/zIO5nGBYDUsEedwXhuxtks4XcZAV0XQhBVl3/NwcXCcExkumfyxB6pRwaTSYDNl+ACzpQiNVH9EjT+s= root@server.corp18.un
---
apiVersion: apps/v1
kind: ReplicaSet
metadata:
  name: my-openssh-server
spec:
  selector:
    matchLabels:
      app: my-openssh-server
  template:
    metadata:
      labels:
        app: my-openssh-server
    spec:
      containers:
      - name: my-openssh-server
        image: linuxserver/openssh-server
        command: ["/bin/sh"]
        args: ["-c", "/usr/bin/ssh-keygen -A; usermod -p '*' root; /usr/sbin/sshd.pam -D"]
        ports:
        - containerPort: 22
        volumeMounts:
        - name: ssh-volume
          subPath: sshd_config
          mountPath: /etc/ssh/sshd_config
        - name: ssh-volume
          subPath: authorized_keys
          mountPath: /root/.ssh/authorized_keys
      volumes:
      - name: ssh-volume
        configMap:
          name: ssh-config
---
apiVersion: v1
kind: Service
metadata:
  name: my-openssh-server
spec:
  type: NodePort
  ports:
  - port: 22
    nodePort: 32222
  selector:
    app: my-openssh-server

# Создаем файл my-openssh-server-deployment.yaml, куда прокидываем ключ сервера и configMap созданный команной выше
# Применяем его
kubectl apply -f my-openssh-server-deployment.yaml
# configmap/ssh-config created
# replicaset.apps/my-openssh-server created
# service/my-openssh-server created

iptables-save | grep 32222
# Теперь подключаясь к порту 32222 пода, мы будем попадать на 22 порт сервиса my-openssh-service
# Можно подключиться с сервера
ssh -p 32222 node2
```

### Установка GOlang
```bash
apt install golang-go
mkdir gowebd
cd gowebd
nano main.go
```
```go
package main

import (
        "fmt"
        "log"
        "net/http"
        "os"
)

const ver = "ver1.1"

func main() {
        http.HandleFunc("/", HelloServer)
        fmt.Printf("Starting server at port 80\n")
        if err := http.ListenAndServe(":80", nil); err != nil {
                log.Fatal(err)
        }
}

func HelloServer(w http.ResponseWriter, r *http.Request) {
        name, err := os.Hostname()
        if err != nil {
                panic(err)
        }
        fmt.Fprint(w, "Hello world from " + name + " " + ver + "\n")
}
```
```bash
# runtime запуск
sudo go run main.go

# сборка двоичного файла
sudo go build -o /usr/local/sbin/gowebd
# Запуск
sudo /usr/local/sbin/gowebd

# Ставим docker
apt install docker.io
usermod -aG docker student
su - student
# student/password

# Создаем Dockerfile
```
```Dockerfile
FROM golang
#FROM golang as builder

WORKDIR /build
COPY . .
RUN test -e go.mod || go mod init gowebd

#ENV CGO_ENABLED=0
RUN go build -o /gowebd

#FROM alpine
#COPY --from=builder /gowebd /gowebd

ENTRYPOINT ["/gowebd"]
```

```bash
# Сборка
time docker build -t gowebd .
0
# Однако, это создаст образ размером под 800мб
# Выход - Multistage Building
```
```Dockerfile
# Из образа голанг создаем билдер
FROM golang as builder

ARG CGO_ENABLED=0
ARG GOOS=linux

# Внутри билдера собирается приложение
WORKDIR /build
COPY . .
RUN test -e go.mod || go mod init gowebd

# Удаляем отладочную информацию из бинарника
RUN go build -ldflags "-s -w" -o /gowebd 

# А потом берем образ Alpine и собранное приложение в билдере копируем в этот образ
FROM alpine
COPY --from=builder /gowebd /gowebd

ENTRYPOINT ["/gowebd"]
```bash
time docker build --progress=plain --no-cache -t gowebd .
```

### Кто слушает порт?
```bash
ss -tupan | grep 22
```

### Работаем с GitLab
```bash
# По умолчанию, Gitlab из контейнера работает во вне на порту 2222, нужно сконфигурировать SSH
nano ./ssh/config

# Вставляем туда
Host server* 
        Port 2222

# GitLab docker registry
# Превращаем GitLab (запущенный в docker) в хранилище docker репозиториев
nano /etc/gitlab/gitlab.rb

# Прописываем
registry_external_url 'http://server.corp18.un'
gitlab_rails['registry_enabled'] = true
gitlab_rails['registry_host'] = "server.corp18.un"
gitlab_rails['registry_port'] = "5000"
registry['registry_http_addr'] = "server.corp18.un:5000"

# Далее
docker exec -it root_web_1 bash
time gitlab-ctl reconfigure

# Настраиваем связку с docker registry (на компьютере разработчика client1 и gate)
nano /etc/docker/daemon.json

{
  "insecure-registries" : ["server.corp18.un:5000"]
}

service docker restart

# логинимся под student
# GitLab -> Deploy -> Container Registry
docker login server.corp18.un:5000 # student/Pa$$w0rd
#docker build -t server.corp18.un:5000/student/gowebd .

docker tag gowebd server.corp18.un:5000/student/gowebd
docker push server.corp18.un:5000/student/gowebd
# Проверили добавление образа и удаляем его, поскольку будет собирать в гитлабе

# Теперь нужно создать персональный gitlab-runner под проект (на gate)
# registration-token берем из GitLab -> Settings -> CI/CD -> Runners и нажимаем на три точки рядом с New Project Runner
# Копируем токен
docker run --rm -v /srv/gitlab-runner/config:/etc/gitlab-runner gitlab/gitlab-runner register \
  --non-interactive \
  --url "http://server.corp18.un/" \
  --registration-token "GR1348941dZtdxS8wHKKdz-xUTAWH" \
  --executor "docker" \
  --docker-image "docker:stable" \
  --docker-privileged \
  --description "gowebd-runner"

# На клиенте создаем файл для пайплайна
nano gowebd/.gitlab-ci.yml
```
```yaml
stages:
  - build
  - push

variables:
  DOCKER_TLS_CERTDIR: ""

services:
  - name: docker:dind
    command:
      [
        '--insecure-registry=server.corp18.un:5000',
      ]

before_script:
  - env
  - echo -n $CI_REGISTRY_PASSWORD | docker login -u $CI_REGISTRY_USER --password-stdin $CI_REGISTRY

Build:
  stage: build
  script:
    - docker pull $CI_REGISTRY_IMAGE:latest || true
    - >
      docker build
      --pull
      --cache-from $CI_REGISTRY_IMAGE:latest
      --tag $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
      .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA

Push latest:
  variables:
    GIT_STRATEGY: none
  stage: push
  only:
    - main
  script:
    - docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker tag $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA $CI_REGISTRY_IMAGE:latest
    - docker push $CI_REGISTRY_IMAGE:latest

Push tag:
  variables:
    GIT_STRATEGY: none
  stage: push
  only:
    - tags
  script:
    - docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker tag $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA $CI_REGISTRY_IMAGE:$CI_COMMIT_REF_NAME
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_REF_NAME
```
```bash
nano main.go # меняем версию
git add . -v # скрытые файлы учитывать
git commit -m "Upgrate to ver1.2"
git tag ver1.2
git push ver1.2

# Переносим конфигурацию insecure registry в GitLab
# удаляем на gate
rm /etc/docker/daemon.json
docker stop gitlab-runner
docker rm gitlab-runner
rm /srv/gitlab-runner/config/config.toml

# Удалить все, включая volume
docker system prune -a --volumes

docker run -d --name gitlab-runner --restart always \
  -v /srv/gitlab-runner/config:/etc/gitlab-runner \
  -v /var/run/docker.sock:/var/run/docker.sock \
  gitlab/gitlab-runner:latest

# Docker in Docker
docker run --rm -v /srv/gitlab-runner/config:/etc/gitlab-runner gitlab/gitlab-runner register \
  --non-interactive \
  --url "http://server.corp18.un/" \
  --registration-token "GR1348941dZtdxS8wHKKdz-xUTAWH" \
  --executor "docker" \
  --docker-image "docker:stable" \
  --docker-privileged \
  --description "dind-runner"

# Обновляем пайплайн для DinD
```
```yaml
stages:
  - build
  - push
#  - deploy

variables:
  DOCKER_TLS_CERTDIR: ""

services:
  - name: docker:dind
    command:
      [
        '--insecure-registry=server.corp18.un:5000',
      ]

before_script:
  - env
  - docker info
  - echo -n $CI_REGISTRY_PASSWORD | docker login -u $CI_REGISTRY_USER --password-stdin $CI_REGISTRY

Build:
  stage: build
#  image:
#    name: gcr.io/kaniko-project/executor:v1.9.0-debug
#    entrypoint: [""]
  script:
    - docker pull $CI_REGISTRY_IMAGE:latest || true
    - >
      docker build
      --pull
      --cache-from $CI_REGISTRY_IMAGE:latest
      --tag $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
      .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA

#    - echo "{\"auths\":{\"${CI_REGISTRY}\":{\"auth\":\"$(printf "%s:%s" "${CI_REGISTRY_USER}" "${CI_REGISTRY_PASSWORD}" | base64 | tr -d '\n')\"},\"$CI_DEPENDENCY_PROXY_SERVER\":{\"auth\":\"$(printf "%s:%s" ${CI_DEPENDENCY_PROXY_USER} "${CI_DEPENDENCY_PROXY_PASSWORD}" | base64 | tr -d '\n')\"}}}" > /kaniko/.docker/config.json
#    - /kaniko/executor
#      --insecure --skip-tls-verify
#      --context "${CI_PROJECT_DIR}"
#      --dockerfile "${CI_PROJECT_DIR}/Dockerfile"
#      --destination "${CI_REGISTRY_IMAGE}:${CI_COMMIT_SHA}"

Push latest:
#  image:
#    name: gcr.io/go-containerregistry/crane:debug
#    entrypoint: [""]
  variables:
    GIT_STRATEGY: none
  stage: push
  only:
    - main
  script:
    - docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker tag $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA $CI_REGISTRY_IMAGE:latest
    - docker push $CI_REGISTRY_IMAGE:latest

#    - crane auth login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
#    - crane --insecure cp $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA $CI_REGISTRY_IMAGE:latest

Push tag:
#  image:
#    name: gcr.io/go-containerregistry/crane:debug
#    entrypoint: [""]
  variables:
    GIT_STRATEGY: none
  stage: push
  only:
    - tags
  script:
    - docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker tag $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA $CI_REGISTRY_IMAGE:$CI_COMMIT_REF_NAME
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_REF_NAME

#   - crane auth login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
#   - crane --insecure cp $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA $CI_REGISTRY_IMAGE:$CI_COMMIT_REF_NAME

#Deploy:
#  variables:
#    VER: "$CI_COMMIT_REF_NAME"
#  stage: deploy
#  only:
#    - tags
#  trigger:
#    project: student/gowebd-k8s
```