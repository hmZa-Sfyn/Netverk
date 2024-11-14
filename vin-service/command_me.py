from flask import Flask, jsonify, request, abort
import sys
import logging
import subprocess
from functools import wraps
import requests

def test_run_command(BASE_URL,COMMAND,BAREER):
        """Test the /f/ssh/cmd endpoint."""
        #print("Testing /f/ssh/cmd...")
        command = COMMAND
        response = requests.post(
            f"{BASE_URL}/f/ssh/cmd",
            headers={"Authorization": f"Bearer {BAREER}"},
            json={"command": command},
        )
        ##self.assertEqual(response.status_code, 200, f"Failed: {response.status_code} - {response.text}")
        print(f"{str(response.json())}")


test_run_command(sys.argv[1],sys.argv[2],sys.argv[3])