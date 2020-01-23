#! /bin/bash

if [[ "$PWD" =~ "go/src/github.com/mobiledgex" ]]; then
  echo "install_edgebox.sh can not be invoked from $PWD. Please stage it under a directory which is not under $GOPATH and invoke it from there."
  exit
fi

# If getting_started_vars.yml is not there, create it
if [[ ! -f getting_started_vars.yml ]]; then
cat <<EOF >getting_started_vars.yml
golang_package_cache: /tmp/golang
golang_version: 1.12.13
golang_checksum: sha256:6d3de6f7d7c0e8162aaa009128839fa5afcba578dcbd6ff034a82419d82480e9 
EOF
fi 

cat <<EOF >inventory
[localhost]
127.0.0.1  ansible_connection=local
EOF

# Check if brew is installed
which brew &> /dev/null  
if [[ ! $? -eq 0  ]]; then
    echo brew not installed, installing it 
    # echo /usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
    /usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"
else
    echo brew is installed
fi

#Check if pip is installed
which pip &> /dev/null
if [[ ! $? -eq 0  ]]; then
#    echo pip not installed, installing it
#    curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py
#    python get-pip.py
    brew install python@2
else
    echo pip is installed
fi

# Check if Ansible is installed
which ansible &> /dev/null
if [[ ! $? -eq 0  ]]; then
    echo ansible not installed, installing it
    brew install ansible
else
    echo ansible is installed
fi

# install golang if needed
which go &> /dev/null
if [[ ! $? -eq 0  ]]; then
    echo go not installed, installing it. Supply root password if you are prompted for password
    sudo ansible-playbook  golang.yml 
else
    echo go is installed
fi

# Generate brew.yml

cat <<EOF >brew.yml
- hosts: localhost
  vars_files:
    - getting_started_vars.yml
  
  tasks: 
  - homebrew:
      name:  "{{ item }}"
      state: present
    loop: 
      - protobuf
      - etcd
      - influxdb
      - postgres
      - vault
      - wget
      - md5sha1sum
      - kubernetes-helm
      - git

EOF

# Install brew packages
ansible-playbook -i inventory brew.yml

# Generate pip.yml
cat <<EOF >pip.yml
- hosts: localhost
  vars_files:
    - getting_started_vars.yml
  
  tasks: 
  - name: Install python openstackclient, heatclient and gnocchiclient using pip
    pip:
      name:  "{{ item }}"
      state: present
    loop: 
      - python-openstackclient 
      - python-heatclient
      - gnocchiclient
EOF

# Install pip packages
echo "Invoking pip.yml ansible playbook."
ansible-playbook -i inventory pip.yml

# Generate git.yml
cat <<EOF >git.yml
- hosts: localhost
  vars_files:
    - getting_started_vars.yml
  
  tasks: 
  - name: Create backup of pre-existing edge-cloud-infra, edge-cloud and edge-proto directories. 
    archive:
      path:
        - ~/go/src/github.com/mobiledgex/edge-proto
        - ~/go/src/github.com/mobiledgex/edge-cloud
        - ~/go/src/github.com/mobiledgex/edge-cloud-infra
      dest: ~/edge-cloud-repos-backup.{{ '%Y-%m-%d %H:%M:%S' | strftime(ansible_date_time.epoch) }}.tgz
      format: zip

  - name: Remove previous edge-cloud, edge-cloud-infra and edge-proto directories
    shell:
      cmd: "/bin/rm -rf {{ item }}"
      warn: false
    loop:
      - ~/go/src/github.com/mobiledgex/edge-proto
      - ~/go/src/github.com/mobiledgex/edge-cloud
      - ~/go/src/github.com/mobiledgex/edge-cloud-infra

  - name: Clone edge-cloud, edge-cloud-infra and edge-proto directories
    git:
      name:  "{{ item.name }}"
      dest: "~/go/src/github.com/mobiledgex/{{ item.handle }}"
      clone: yes
      update: yes
    loop: 
      - { name: 'https://github.com/mobiledgex/edge-cloud.git', handle: 'edge-cloud' }
      - { name: 'https://github.com/mobiledgex/edge-cloud-infra.git', handle: 'edge-cloud-infra' }
      - { name: 'https://github.com/mobiledgex/edge-proto.git', handle: 'edge-proto' }

  - name: Run go mod download and make tools in edge-cloud directory
    shell:
      cmd:  "cd ~/go/src/github.com/mobiledgex/edge-cloud; GO111MODULE=on go mod download; make tools"

  - name: Run make , make unit-test  in edge-cloud-infra directory
    shell:
      cmd:  "cd ~/go/src/github.com/mobiledgex/edge-cloud-infra;  make ;  make unit-test"

  - name: Run  make test in edge-cloud-infra directory. This takes time as there are lots of tests.
    shell:
      cmd:  "cd ~/go/src/github.com/mobiledgex/edge-cloud-infra;   make test"

  - name: Run make install-dind in edge-cloud directory
    shell:
      cmd:  "cd ~/go/src/github.com/mobiledgex/edge-cloud; make install-dind"
EOF
# Install git packages
echo "Invoking git.yml ansible playbook."
ansible-playbook -i inventory git.yml

