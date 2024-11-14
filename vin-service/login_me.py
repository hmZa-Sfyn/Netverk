from flask import Flask, jsonify, request, abort
import sys
import logging
import subprocess
from functools import wraps
import requests

def test_login(BASE_URL,USERNAME,PASSWORD):
        """Test the /f/auth/login endpoint and return the auth token."""
        #print("Testing /f/auth/login...")
        response = requests.post(
            f"{BASE_URL}/f/auth/login",
            json={"username": USERNAME, "password": PASSWORD},
        )
        if response.status_code == 200:
            auth_token = response.json().get("auth-token")
            print(f"{auth_token}")
            return auth_token
        else:
            print(f"###not_logined###: {response.json()}")

test_login(sys.argv[1],sys.argv[2],sys.argv[3])