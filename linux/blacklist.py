import json
import subprocess
import os.path
import sys


def blacklist():
    try:
        res = subprocess.Popen("psql -qAt -U shift shift_db -c \"SELECT ip FROM peers WHERE version < '5.0.0';\"",  
                            shell=True, stdout=subprocess.PIPE).stdout.read()
    except Exception as e:
        print "Could not fetch IP addresses from SQL server, reason: %s" % e

    if os.path.isfile("config.json"):
        try:
            with open("config.json", "r+") as f:
                content = json.loads(f.read())
                content['peers']['blackList'] = res.split()
                f.seek(0)
                f.write(json.dumps(content, indent=4, sort_keys=True))
                f.truncate()
        except Exception as e:
            print "Could not open configuration file for Shift node. Exiting."
            sys.exit(0)
    return res

if __name__ == "__main__":
    res = blacklist()
    if res:
        print "Blacklisted %i addresses, restart your node." % len(res.split())
    sys.exit(0)
