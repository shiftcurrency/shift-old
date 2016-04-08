#!/usr/bin/env python

import json
import sqlite3
import requests
import sys
import os
import numpy as np
from time import sleep

output_file = "shift_2.5.0.json"

def fetch_accounts():


    ''' Fetch the current blockheight '''
    data = json.dumps({"jsonrpc":"2.0","method":"shf_blockNumber","params":[],"id":83})
    accounts = []

    try:
        response = requests.post("http://localhost:53901", data=data)
        jsondata = response.json()
        current_height = int(jsondata['result'], 16)

    except Exception as e:
        print "Hit a problem with HTTP request."
        print e
        sys.exit(0)

    ''' Fetch all blocks and the accounts that have made a transaction'''
    for num in range(0,current_height):

        blocks = "Parsing block: %i" % num
        print "\r", blocks,

        data_string = '{"jsonrpc":"2.0","method":"shf_getBlockByNumber","params":["%s",true],"id":1}' % str(hex(num))
    
        try:
            sleep(0.005)
            response = requests.post("http://localhost:53901", data=data_string)
            jsondata = response.json()
            if (jsondata['result']['transactions']):
                for i in jsondata['result']['transactions']:
                    accounts.append(i['to'])
                    accounts.append(i['from'])

        except Exception as e:
            print "Hit a problem with HTTP request."
            print e
            sys.exit(0)

    ''' Load the accounts in the old genesis block (These accounts does not show up as transactions)'''
    try:
        data = open("shift_2.4.1.json").read()
        former_genesis = json.loads(data)
    except Exception as e:
        print "Count not open shift_2.4.1.json."
        print e
        sys.exit(0)

    for i in former_genesis['alloc']:
        accounts.append(i)


    ''' Return all unique accounts as a list '''
    return np.unique(accounts)
        


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


    if os.path.isfile(output_file):
        try:
            os.remove(output_file)
            print "Removed old json file."

        except Exception as e:
            print "Could not remove old json file."
            print e

    try:
        with open(output_file, "a") as genfile:

            print "Creating genesis account:balance allocation..."
            genfile.write('{ "nonce": "0x0000000000000042", \n"difficulty": "0x273942957", \n"alloc": {\n')
            for account in account_balances:
                if account != "" and int(account_balances[account]) != 0:
                    account_allocation = " \"%s\": { \"balance\": \"%s\" },\n" % (account, account_balances[account])
                    genfile.write(account_allocation)

            genesis_ending = '},\n "mixhash": "0x0000000000000000000000000000000000000000000000000000000000000000",\n' + \
                             '"coinbase": "0x0000000000000000000000000000000000000000",\n' + \
                             '"timestamp": "0x00",\n "parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",\n' + \
                             '"gasLimit": "0x300000"\n'
            genfile.write(genesis_ending)

        return True

    except Exception as e:
        return False
    



if __name__ == "__main__":
    account_list = fetch_accounts()
    account_balances = get_balances(account_list)
    total = 0
    for i in account_balances:
        total += int(account_balances[i])
    print "Total number of SHIFT: %s" % str(total)[:7]
    print "Total number of accounts: %s" % str(len(account_balances))


    ''' write the final .json file '''
    if create_genesis_json(account_balances):
        print "Done. See the created file %s." % output_file
        
