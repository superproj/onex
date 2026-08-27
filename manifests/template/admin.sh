#!/bin/bash
#
# admin.sh : 简易进程管理脚本
# 支持命令: start | stop | restart | status
# 作者: yourname
#

### 配置区 ##########################################################

# 启动命令（可按需修改）
ExecStart="/usr/bin/python3 -m http.server 8080"

# PID 文件保存路径
PID_FILE="$PWD/admin.pid"

#####################################################################

# 获取进程状态
get_pid() {
    if [[ -f "$PID_FILE" ]]; then
        PID=$(cat "$PID_FILE" 2>/dev/null)
        if [[ -n "$PID" && -d "/proc/$PID" ]]; then
            echo "$PID"
            return 0
        else
            rm -f "$PID_FILE" >/dev/null 2>&1
            return 1
        fi
    else
        return 1
    fi
}

start() {
    echo ">>> 启动服务中..."
    if get_pid >/dev/null; then
        echo "服务已运行，PID=$(cat $PID_FILE)"
        return 0
    fi
    nohup $ExecStart >$PWD/admin.log 2>&1 &
    echo $! > "$PID_FILE"
    sleep 1
    if get_pid >/dev/null; then
        echo "✅ 服务启动成功，PID=$(cat $PID_FILE)"
    else
        echo "❌ 启动失败，请检查日志 $PWD/admin.log"
    fi
}

stop() {
    echo ">>> 停止服务中..."
    if get_pid >/dev/null; then
        PID=$(cat "$PID_FILE")
        kill "$PID" >/dev/null 2>&1 || true
        sleep 1
        if get_pid >/dev/null; then
            echo "⚠️ 进程 $PID 未能停止，尝试强制中止..."
            kill -9 "$PID" >/dev/null 2>&1 || true
        fi
        rm -f "$PID_FILE"
        echo "✅ 服务已停止。"
    else
        echo "服务未在运行，无需停止。"
    fi
}

restart() {
    echo ">>> 重启服务中..."
    stop
    start
}

status() {
    if get_pid >/dev/null; then
        echo "✅ 服务运行中，PID=$(cat $PID_FILE)"
    else
        echo "❌ 服务未运行。"
    fi
}

case "$1" in
    start) start ;;
    stop) stop ;;
    restart) restart ;;
    status) status ;;
    *)
        echo "用法: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac
