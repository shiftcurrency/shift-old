#!/usr/bin/env python

import json
import sqlite3
import requests
import sys
import os

conn = sqlite3.connect(".gshift/sql.db")
c = conn.cursor()

def fetch_accounts():

    try:
        c.execute("SELECT DISTINCT(receiver) FROM shift_transactions ORDER BY receiver;")
        accounts = c.fetchall()
    except Exception as exception:
        print "Hit a problem, see the output."
        print exception
        sys.exit(0)

    account_list = []

    if len(accounts) > 0:
        for i in accounts:
            account_list.append(i[0])
    
    return account_list
        

def get_balances(accounts):

    account_balance = {}
   
    print "Collecting account balances via RPC"

    for account in accounts:
        data = json.dumps({"jsonrpc":"2.0","method":"shf_getBalance","params":["%s", "latest"],"id":1}) % (account)

        try:
            response = requests.post("http://localhost:53901", data=data)
        except Exception as e:
            print "Hit a problem with HTTP request."
            print e
            sys.exit(0)

        jsondata = response.json()
        if 'result' in jsondata:
            ''' Use str to remove trailing L '''
            account_balance[account] = str(int(jsondata['result'], 16))
    
    if not len(account_balance) > 0:
        print "Could not find any accounts with funds on it. Exiting."
        sys.exit(0)

    return account_balance


def create_genesis_json(account_balances):


    if os.path.isfile("shift_2.4.1.json"):
        try:
            os.remove("shift_2.4.1.json")
            print "Removed old json file."

        except Exception as e:
            print "Could not remove old json file."
            print e

    try:
        with open("shift_2.4.1.json", "a") as genfile:

            print "Creating genesis account:balance allocation..."
        
            genfile.write('"alloc": {\n')
            for account in account_balances:
                if account != "" and int(account_balances[account]) != 0:
                    account_allocation = " \"%s\": { \"balance\": \"%s\" },\n" % (account, account_balances[account])
                    genfile.write(account_allocation)
            genfile.write('},\n')

        return True

    except Exception as e:
        return False
    



if __name__ == "__main__":
    account_list = fetch_accounts()
    ''' returns a dictionary with account and balance '''

    account_balances = get_balances(account_list)

    total = 0
    for i in account_balances:
        total += int(account_balances[i])
    print "Total number of SHIFT: %s" % str(total)[:7]
    print "Total number of accounts: %s" % str(len(account_balances))


    ''' write the final .json file '''
    if create_genesis_json(account_balances):
        print "Done. See the created file shift_2.4.1.json."
        
