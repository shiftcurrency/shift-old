#!/usr/bin/env bash
export LC_ALL=en_US.UTF-8
export LANG=en_US.UTF-8
export LANGUAGE=en_US.UTF-8
logfile="shift_installer.log"

if [ "\$USER" == "root" ]; then
  echo "Error: SHIFT should not be installed be as root. Exiting."
  exit 1
fi


install_prereq() {
    sudo id &> /dev/null

    echo -n "Running: apt-get update... ";
    sudo apt-get update  &> /dev/null || \
    { echo "Could not update apt repositories. Run apt-get update manually. Exiting." && exit 1; };
    echo -e "done.\n"

    echo -n "Running: apt-get install curl build-essential python lsb-release wget... ";
    sudo apt-get install -y -qq curl build-essential python lsb-release wget &>> $logfile || \
    { echo "Could not install packages prerequisites. Exiting." && exit 1; };
    echo -e "done.\n"

    echo -n "Removing former postgresql installation... ";
    sudo apt-get purge -y -qq postgres* &>> $logfile || \
    { echo "Could not remove former installation of postgresql. Exiting." && exit 1; };
    echo -e "done.\n"

    echo -n "Updating apt repository sources for postgresql.. ";
    sudo bash -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ wheezy-pgdg main" > /etc/apt/sources.list.d/pgdg.list' &>> $logfile || \
    { echo "Could not add postgresql repo to apt." && exit 1; }
    echo -e "done.\n"

    echo -n "Adding postgresql repo key... "
    sudo wget -q https://www.postgresql.org/media/keys/ACCC4CF8.asc -O - | sudo apt-key add - &>> $logfile || \
    { echo "Could not add postgresql repo key. Exiting." && exit 1; }
    echo -e "done.\n"

    echo -n "Installing postgresql... "
    sudo apt-get update -qq &> /dev/null && sudo apt-get install -y -qq postgresql-9.6 postgresql-contrib-9.6 libpq-dev &>> $logfile || \
    { echo "Could not install postgresql. Exiting." && exit 1; }
    echo -e "done.\n"

    return 0;
}


ntp_checks() {
    # Install NTP or Chrony for Time Management - Physical Machines only
    if [[ ! -f "/proc/user_beancounters" ]]; then
      if ! sudo pgrep -x "ntpd" > /dev/null; then
        echo -n "\nInstalling NTP... "
        sudo apt-get install ntp -yyq &>> $logfile
        sudo service ntp stop &>> $logfile
        sudo ntpdate pool.ntp.org &>> $logfile
        sudo service ntp start &>> $logfile
        if ! sudo pgrep -x "ntpd" > /dev/null; then
          echo -e "SHIFT requires NTP running. Please check /etc/ntp.conf and correct any issues. Exiting."
          exit 1
        echo -e "done.\n"
        fi # if sudo pgrep
      fi # if [[ ! -f "/proc/user_beancounters" ]]
    elif [[ -f "/proc/user_beancounters" ]]; then
      echo -e "Running OpenVZ or LXC VM, NTP is not required, done. \n"
    fi
}

add_pg_user_database() {

    if start_postgres; then
        user_exists=$(grep postgres /etc/passwd |wc -l);
        if [[ $user_exists == 1 ]]; then
            echo -n "Creating database user... "
            res=$(sudo -u postgres psql -c "CREATE USER shift WITH PASSWORD 'testing';" 2> /dev/null)
            res=$(sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='shift'" 2> /dev/null)
            if [[ $res -eq 1 ]]; then
                echo -e "done.\n"
            fi

            echo -n "Creating database... "
            res=$(sudo -u postgres createdb -O shift shiftdb 2> /dev/null)
            res=$(sudo -u postgres psql -lqt 2> /dev/null |grep shiftdb |awk {'print $1'} |wc -l)
            if [[ $res -eq 1 ]]; then
                echo -e "done.\n"
            fi
        fi
        return 0
    fi

    return 1;
}

start_postgres() {

    installed=$(dpkg -l |grep postgresql |grep ii |head -n1 |wc -l);
    running=$(ps aux |grep "bin\/postgres" |wc -l);

    if [[ $installed -ne 1 ]]; then
        echo "Postgres is not installed. Install postgres manually before continuing. Exiting."
        exit 1;
    fi

    if [[ $running -ne 1 ]]; then
        sudo /etc/init.d/postgresql start &>> $logfile || { echo "Could not start postgresql, try to start it manually. Exiting." && exit 1; }
    fi

    #with systemd the service is not automatically enabled on restart
    if [[ `sudo systemctl` =~ -\.mount ]]; then
        sudo systemctl enable postgresql &>> $logfile || { echo "Could not enable postgresql service (systemd). Exiting." && exit 1; }
    fi

    return 0
}

install_node_npm() {

    echo -n "Installing nodejs and npm... "
    curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash - &>> $logfile
    sudo apt-get install -y -qq nodejs &>> $logfile || { echo "Could not install nodejs and npm. Exiting." && exit 1; }
    echo -e "done.\n" && echo -n "Installing grunt-cli... "
    sudo npm install grunt-cli -g &>> $logfile || { echo "Could not install grunt-cli. Exiting." && exit 1; }
    echo -e "done.\n" && echo -n "Installing bower... "
    sudo npm install bower -g &>> $logfile || { echo "Could not install bower. Exiting." && exit 1; }
    echo -e "done.\n"

    return 0;
}

install_shift() {

    echo -n "Installing Shift core... "
    npm install --production &>> $logfile || { echo "Could not install SHIFT, please check the log directory. Exiting." && exit 1; }
    echo -e "done.\n"

    return 0;
}

install_webui() {

    echo -n "Installing Shift WebUi... "
    git clone https://github.com/shiftcurrency/shift-wallet &>> $logfile || { echo -n "Could not clone git wallet source. Exiting." && exit 1; }

    if [[ -d "public" ]]; then
        rm -rf public/
    fi

    if [[ -d "shift-wallet" ]]; then
        mv shift-wallet public
    else
        echo "Could not find installation directory for SHIFT web wallet. Install the web wallet manually."
        exit 1;
    fi

    cd public && npm install &>> $logfile || { echo -n "Could not install web wallet node modules. Exiting." && exit 1; }
    # Bower config seems to have the wrong permissions. Make sure we change these before trying to use bower.
    sudo chown -R $USER:$USER ~/.config &> /dev/null
    bower install &>> $logfile || { echo -n "Could not install bower components for the web wallet. Exiting." && exit 1; }
    grunt release &>> $logfile || { echo -n "Could build web wallet release. Exiting." && exit 1; }
    echo "done."
    
    return 0;

}

parse_option() {
  OPTIND=2
  while getopts d:r:n opt
  do
    case $opt in
      u) install_with_ui=true ;;
    esac
  done
}


case $1 in
"install")
parse_option $@
install_prereq
ntp_checks
add_pg_user_database
install_node_npm
install_shift
install_webui
  ;;
*)
  echo "Usage: ./shift_installer.bash install"
  exit 1
  ;;
esac

echo ""
echo ""
echo ""
echo "Start SHIFT with: node app.js"
echo "Open the User Interface with: http://node.ip:8005"

exit 0;
