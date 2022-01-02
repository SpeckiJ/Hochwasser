#!/bin/sh

# sets up BBR congestion control (https://git.kernel.org/pub/scm/linux/kernel/git/netdev/net-next.git/commit/?id=0f8782ea14974ce992618b55f0c041ef43ed0b78)
# for the given interface

DEV=${$1:=eth0}
MODE=${$2:="temporary"}

echo pre-run state:
echo `sysctl net.ipv4.tcp_congestion_control`
echo qdisc on $DEV: `tc qdisc show dev $DEV`

# set required "fair queueing" queuing discipline (man tc-fq)
echo "qdisc on $DEV was:     `tc qdisc show dev $DEV`"
tc qdisc replace dev $DEV root fq
echo "qdisc on $DEV now is:  `tc qdisc show dev $DEV`"

# (permanently) configure BBR
if [[ mode == "temporary" ]]; then
    echo -e 'net.core.default_qdisc=fq
    net.ipv4.tcp_congestion_control=bbr' | sysctl -p-
elif [[ mode == "permanent"]]; then
    echo -e 'net.core.default_qdisc=fq
    net.ipv4.tcp_congestion_control=bbr' >> /etc/sysctl.d/20-net-bbr.conf
    sysctl -p /etc/sysctl.d/20-net-bbr.conf
    algo=`sysctl net.ipv4.tcp_congestion_control`
    if [[ "$algo" != 'net.ipv4.tcp_congestion_control = bbr' ]]; then
        echo "couldn't apply changes at runtime, you should reboot."
    fi
else
    echo "unknown mode, congestion control algo unchanged"
fi
