#!/bin/bash

SERVICE_ENABLE=0

while getopts "s:" opt; do
  case $opt in
    s)
      SERVICE_ENABLE=1
      echo "Enabling services"
      ;;
    \?) # Handle invalid options
      echo "Building app"
      ;;
  esac
done
setup_logo_service_T(){
  sudo ls /etc/systemd/system
  echo "DIR: $1"
  printf "service: %s\n" "$2"
}

setup_logo_service(){
    PLC_DIR="$1"
    SERV_NAME="$2"
    APP_DIR="$HOME/$PLC_DIR"
    if [ "$SERVICE_ENABLE" -eq 0 ]; then
      sudo systemctl stop "$SERV_NAME.service"

      mkdir -p "$APP_DIR"
      cp start.sh "$APP_DIR/"
      cd ".."
      #go build -o "$APP_DIR/bin" .
      cp bin-arm "$APP_DIR/bin"
      cp ".env.$SERV_NAME" "$APP_DIR/.env"

      cd "$APP_DIR" || exit
      sudo chmod 775 start.sh
      sed -i "s/DIR/$PLC_DIR/g" start.sh
      sudo systemctl enable "$SERV_NAME.service"
      sudo systemctl start "$SERV_NAME.service"
    else
      cp logo.service "$APP_DIR/"
      cd "$APP_DIR" || exit
      sed -i "s/{DIR}/$PLC_DIR/g" logo.service
      sudo cp logo.service "/etc/systemd/system/$SERV_NAME.service"
      sudo systemctl daemon-reload
      sudo systemctl enable "$SERV_NAME.service"
      sudo systemctl restart "$SERV_NAME.service"
      echo "Service $SERV_NAME started"
    fi
}


source env.sh
readarray -t dirs <<< "$LOGO_DIRS"
readarray -t services <<< "$LOGO_SERVICES"

for i in "${!dirs[@]}"; do
    dir="${dirs[i]}"
    ser="${services[i]}"
    setup_logo_service "$dir" "$ser"
done
