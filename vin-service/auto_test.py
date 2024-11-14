import unittest
import requests
import json
import time

# Set the API base URL
BASE_URL = "http://127.0.0.1:7009"  # Change this if needed

# User credentials
USERNAME = "root"
PASSWORD = "toor"


class APITestCase(unittest.TestCase):
    
    @classmethod
    def setUpClass(cls):
        """This runs once before all tests to authenticate and get the auth token."""
        print("Running setUpClass - Authenticate to get token")
        cls.auth_token = cls.test_login()

    @staticmethod
    def test_login():
        """Test the /f/auth/login endpoint and return the auth token."""
        print("Testing /f/auth/login...")
        response = requests.post(
            f"{BASE_URL}/f/auth/login",
            json={"username": USERNAME, "password": PASSWORD},
        )
        if response.status_code == 200:
            auth_token = response.json().get("auth-token")
            print(f"Login successful! Auth token: {auth_token}")
            return auth_token
        else:
            raise AssertionError(f"Login failed: {response.status_code} - {response.text}")

    def test_hello(self):
        """Test the /hello endpoint."""
        print("Testing /hello...")
        response = requests.get(f"{BASE_URL}/hello")
        self.assertEqual(response.status_code, 200, f"Failed: {response.status_code} - {response.text}")
        print(f"Response: {response.json()}")

    def test_routes(self):
        """Test the /routes endpoint."""
        print("Testing /routes...")
        response = requests.get(f"{BASE_URL}/routes")
        self.assertEqual(response.status_code, 200, f"Failed: {response.status_code} - {response.text}")
        print(f"Available routes: {response.json()}")

    def test_run_command(self):
        """Test the /f/ssh/cmd endpoint."""
        print("Testing /f/ssh/cmd...")
        command = "dir"
        response = requests.post(
            f"{BASE_URL}/f/ssh/cmd",
            headers={"Authorization": f"Bearer {self.auth_token}"},
            json={"command": command},
        )
        self.assertEqual(response.status_code, 200, f"Failed: {response.status_code} - {response.text}")
        print(f"Command response: {response.json()}")

    @staticmethod
    def test_shutdown():
        url = "http://127.0.0.1:7009/f/system/shutdown"
        headers = {"Authorization": f"Bearer {test_login()}"}
        
        response = requests.post(url, headers=headers)
        
        if response.status_code == 200:
            print("Shutdown successful.")
        else:
            print(f"Failed: {response.status_code} - {response.json()}")



if __name__ == "__main__":
    # Run the tests
    unittest.main()
