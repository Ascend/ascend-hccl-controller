# Copyright © Huawei Technologies Co., Ltd. 2020. All rights reserved.
---
# Generate report with master
- hosts: localnode, master
  remote_user: root

  tasks:
    - name: Change file mode
      file:
        path: "{{ playbook_dir }}/check_env.sh"
        mode: 0540
      ignore_errors: True

    - name: Check MindX-DL env - master nodes
      shell:
        chdir: "{{ playbook_dir }}"
        cmd:
          dos2unix check_env.sh;
          bash check_env.sh master "{{ ansible_default_ipv4['address'] }}" >> /dev/null 2>&1
      when:
        - ansible_hostname not in groups["workers"]

    - name: Check MindX-DL env - master-worker nodes
      shell:
        chdir: "{{ playbook_dir }}"
        cmd:
          dos2unix check_env.sh;
          bash check_env.sh master-worker "{{ ansible_default_ipv4['address'] }}" >> /dev/null 2>&1
      when:
        - ansible_hostname in groups["workers"]

- hosts: master
  remote_user: root

  tasks:
    - name: Set reports dir facts
      add_host:
        name: "master_reports"
        report_dir: "{{ playbook_dir }}/reports"

    - name: Remove result dir
      file:
        path: "{{ hostvars['master_reports']['report_dir'] }}"
        state: absent

    - name: Create result dir
      file:
        path: "{{ hostvars['master_reports']['report_dir'] }}"
        state: directory
        mode: 0750

    - name: Create master report txt
      copy:
        src: "{{ playbook_dir }}/env_check_report.txt"
        dest: "{{ hostvars['master_reports']['report_dir'] }}/env_check_report_master.txt"
        remote_src: True
        mode: 0444

    - name: Remove temporary file on master
      shell:
        cmd:
          rm -rf {{ playbook_dir }}/env_check_report.txt;
          rm -rf {{ playbook_dir }}/check_evn.retry
      ignore_errors: True

# Generate report with worker
- hosts: workers
  remote_user: root
  vars:
    tmp_dir: "/tmp"

  tasks:
    - name: Granting Permissions to the Worker Node
      shell:
        cmd:
          mv $HOME/.kube $HOME/.kube.bak || true;
          mkdir $HOME/.kube;
          cp /etc/kubernetes/kubelet.conf $HOME/.kube/config;
      when:
        - ansible_default_ipv4['address'] != master_ip

    - name: Copy check_env.sh to workers
      copy:
        src: "{{ playbook_dir }}/check_env.sh"
        dest: "{{ tmp_dir }}/check_env.sh"
        mode: 0540
      when:
        - ansible_default_ipv4['address'] != master_ip

    - name: Check MindX-DL env - worker nodes
      shell:
        chdir: "{{ tmp_dir }}"
        cmd:
          dos2unix check_env.sh;
          unset KUBECONFIG;
          bash check_env.sh worker "{{ ansible_default_ipv4['address'] }}" >> /dev/null 2>&1
      when:
        - ansible_default_ipv4['address'] != master_ip

    - name: Revoking Permissions on a Worker Node
      shell:
        cmd:
          rm -rf $HOME/.kube || true;
          mv $HOME/.kube.bak $HOME/.kube || true;
          source /etc/profile
      args:
        executable: "/bin/bash"
      when:
        - ansible_default_ipv4['address'] != master_ip

    - name: Create workers report txt
      copy:
        src: "{{ tmp_dir }}/env_check_report.txt"
        dest: "{{ tmp_dir }}/env_check_report_{{ ansible_default_ipv4['address'] }}.txt"
        remote_src: True
        mode: 0444
      when:
        - ansible_default_ipv4['address'] != master_ip

    # copy from worker node
    - name: Get report from workers
      fetch:
        src: "{{ tmp_dir }}/env_check_report_{{ ansible_default_ipv4['address'] }}.txt"
        dest: "{{ hostvars['master_reports']['report_dir'] }}/env_check_report_{{ ansible_default_ipv4['address'] }}.txt"
        flat: True
      when:
        - ansible_default_ipv4['address'] != master_ip

    - name: Remove temporary file on workers
      shell:
        cmd:
          rm -rf {{ tmp_dir }}/check_env.sh;
          rm -rf {{ tmp_dir }}/env_check_report.txt;
          rm -rf {{ tmp_dir }}/env_check_report_{{ ansible_default_ipv4['address'] }}.txt;
      ignore_errors: True
      when:
        - ansible_default_ipv4['address'] != master_ip

# Generate final report
- hosts: master
  remote_user: root
  vars:
    report_file_name: "env_check_report_all.txt"

  tasks:
    - name: Remove old report
      file:
        path: "{{ playbook_dir }}/{{ report_file_name }}"
        state: absent
      ignore_errors: True

    - name: Create new report
      file:
        path: "{{ playbook_dir }}/{{ report_file_name }}"
        state: touch
        mode: 0444

    - name: Generate final report
      shell:
        cmd:
          cat {{ item }} >> {{ playbook_dir }}/{{ report_file_name }};
          echo "" >> {{ playbook_dir }}/{{ report_file_name }}
      with_fileglob: "{{ hostvars['master_reports']['report_dir'] }}/*"
      register: all_reports

    - name: Print report path
      shell:
        cmd:
          echo "Finished! The check report is stored in the {{ playbook_dir }}/{{ report_file_name }} on the master node."