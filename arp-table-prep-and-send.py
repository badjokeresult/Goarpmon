#!/usr/bin/env python3

import json
import subprocess
import sys


def calc_multiple_ips_on_single_mac(arp_table):
    multiple_ips = dict()
    for row in arp_table:
        mac = row[1]
        multiple_ips[mac] = []
        for column in arp_table:
            if column[1] == mac:
                multiple_ips[mac].append(column[0])
    return multiple_ips


def calc_multiple_macs_on_single_ip(arp_table):
    multiple_macs = dict()
    for row in arp_table:
        ip = row[0]
        multiple_macs[ip] = []
        for column in arp_table:
            if column[0] == ip:
                multiple_macs[ip].append(column[1])
    return multiple_macs


def get_jsoned_arp_table(arp_table):
    jsoned = {"data": []}
    for row in arp_table:
        jsoned["data"].append({
            "ipAddress": row[0],
            "hwAddress": row[1],
        })
    return json.dumps(jsoned)


def execute_sending_data_to_zabbix(sender, config, key, data):
    command = subprocess.run([sender, '-c', config, '-k', key, '-o', data, '-vv'])

    if command.returncode != 0:
        raise IOError


def execute_snmp_arp_table_capture():
    arp_table = []
    for line in sys.stdin:
        tmp_line = line.strip()
        addrs = tmp_line.split('=')
        ip = addrs[0].strip()
        mac = ":".join(addrs[1].split(":")[1].strip().split(" "))
        arp_table.append((ip, mac))
    return arp_table


def main():
    arp_table = execute_snmp_arp_table_capture()

    multiple_macs_on_single_ip = calc_multiple_macs_on_single_ip(arp_table)
    multiple_ips_on_single_mac = calc_multiple_ips_on_single_mac(arp_table)

    jsoned_table = get_jsoned_arp_table(arp_table)

    sender = "/usr/bin/zabbix_sender"
    config = "/etc/zabbix/zabbix_agentd.conf"
    execute_sending_data_to_zabbix(sender, config, "arp.discovery", jsoned_table)

    for ip, macs_list in multiple_macs_on_single_ip.items():
        macs_list_as_str = ' '.join(macs_list)
        execute_sending_data_to_zabbix(sender, config, f"arp.ipsMac[{ip}]", macs_list_as_str)

    for mac, ips_list in multiple_ips_on_single_mac.items():
        ips_list_as_str = ' '.join(ips_list)
        execute_sending_data_to_zabbix(sender, config, f"arp.macIps[{mac}]", ips_list_as_str)

    for mac, ips_list in multiple_ips_on_single_mac.items():
        execute_sending_data_to_zabbix(sender, config, f"arp.ipCount[{mac}]", str(len(ips_list)))

    for ip, macs_list in multiple_macs_on_single_ip.items():
        execute_sending_data_to_zabbix(sender, config, f"arp.macCount[{ip}]", str(len(macs_list)))


if __name__ == "__main__":
    main()
