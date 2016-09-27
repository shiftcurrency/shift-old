#!/usr/bin/env bash


install_prereq() {
    sudo apt-get update >> /dev/null || \
    { echo -n "Could not update apt repositories. Run apt-get update manually. Exiting." && exit 1; };
    
    echo -n "Running: apt-get install curl build-essential python...";
    sudo apt-get install -y -qq curl build-essential python 2>&1|| \
    { echo -n "Could not install packages prerequisites. Exiting." && exit 1; };
    echo "done."

    echo -n "Removing former postgresql installations: apt-get purge -y postgres*...";
    sudo apt-get purge -y -qq postgres* >> /dev/null || \
    { echo -n "Could not remove former installation of postgresql. Exiting." && exit 1; };
    echo "done."

    
    echo -n "Updating apt repository sources for postgresql..";
    sudo bash -c 'echo "deb http://apt.postgresql.org/pub/repos/apt/ wheezy-pgdg main" > /etc/apt/sources.list.d/pgdg.list' 2> /dev/null || \
    { echo -n "Could not add postgresql repo to apt." && exit 1; }
    echo "done."
    
    echo -n "Adding postgresql repo key..."
    sudo wget -q https://www.postgresql.org/media/keys/ACCC4CF8.asc -O - | sudo apt-key add - >> /dev/null || \
    { echo -n "Could not add postgresql repo key. Exiting." && exit 1; }
    echo "done."

    echo -n "Installing postgresql..."
    sudo apt-get update -qq >> /dev/null && sudo apt-get install -y -qq postgresql postgresql-contrib libpq-dev 2> /dev/null || \
    { echo -n "Could not install postgresql. Exiting." && exit 1; }
    echo "done."

    return 0;
}

add_pg_user_database() {

    if start_postgres; then
        user_exists=$(grep postgres /etc/passwd |wc -l);
        if [[ $user_exists == 1 ]]; then
            echo -n "Creating database user..."
            res=$(sudo -u postgres psql -c "CREATE USER shift WITH PASSWORD 'testing';" 2> /dev/null)
            res=$(sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='shift'" 2> /dev/null)
            if [[ $res -eq 1 ]]; then
                echo "done."
            fi

            echo -n "Creating database..."
            res=$(sudo -u postgres createdb -O shift shiftdb 2> /dev/null)
            res=$(sudo -u postgres psql -lqt 2> /dev/null |grep shiftdb |awk {'print $1'} |wc -l)
            if [[ $res -eq 1 ]]; then
                echo "done."
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
        /etc/init.d/postgresql start || { echo -n "Could not start postgresql, try to start it manually. Exiting." && exit 1; }
    fi

    return 0
}

install_node_npm() {

    echo -n "Installing nodejs and npm..."
    curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash - >> /dev/null
    sudo apt-get install -y -qq nodejs >> /dev/null || { echo -n "Could not install nodejs and npm. Exiting." && exit 1; }
    echo "done." && echo -n "Installing grunt-cli..."
    sudo npm install grunt-cli -g 2> /dev/null || { echo -n "Could not install grunt-cli. Exiting." && exit 1; }
    echo "done." && echo -n "Installing bower..."
    sudo npm install bower -g 2> /dev/null || { echo -n "Could not install bower. Exiting." && exit 1; }
    echo "done."

    return 0;
}

install_shift() {

    echo -n "Installing SHIFT core..."
    npm install --production 2> /dev/null || { echo -n "Could not install SHIFT, please check the log directory. Exiting." && exit 1; }
    echo "done."
    return 0;
}

install_webui() {




}


install_prereq
add_pg_user_database
install_node_npm
install_shift

echo ""
echo ""
echo ""
echo "Start SHIFT with: node app.js".

exit 0;
