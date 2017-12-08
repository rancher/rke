#!/bin/bash -x

if [ "$1" == "kubelet" ]; then
    for i in $(DOCKER_API_VERSION=1.24 ./docker info 2>&1  | grep -i 'docker root dir' | cut -f2 -d:) /var/lib/docker /run /var/run; do
        for m in $(tac /proc/mounts | awk '{print $2}' | grep ^${i}/); do
            if [ "$m" != "/var/run/nscd" ] && [ "$m" != "/run/nscd" ]; then
                umount $m || true
            fi
        done
    done
    mount --rbind /host/dev /dev
    mount -o rw,remount /sys/fs/cgroup 2>/dev/null || true
    for i in /sys/fs/cgroup/*; do
        if [ -d $i ]; then
             mkdir -p $i/kubepods
        fi
    done
    CGROUPDRIVER=$(docker info | grep -i 'cgroup driver' | awk '{print $3}')
    exec "$@" --cgroup-driver=$CGROUPDRIVER
fi

exec "$@"
