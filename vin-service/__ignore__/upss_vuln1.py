import requests
from bs4 import BeautifulSoup

# URL of the login page
login_url = 'https://7340227CAA397F8BFDF9EEE8701FD44379A1DD3F12A36B7B6B205137C7FDEF5C.onion/login'

# Data to be sent in the form (replace with actual email and password)
form_data = {
    '_token': 'pqDXc1G8b4Fd7BmkBIYZgKteTaRBg6Evjb4urK02',  # token used for CSRF protection
    'email': 'your_email@example.com',  # replace with actual email or CNIC
    'password': 'your_password',        # replace with actual password
}

# Simulating the login request
response = requests.post(login_url, data=form_data)

# Check if the login was successful
if response.status_code == 200:
    print("Login successful")
else:
    print(f"Login failed with status code: {response.status_code}")

# Optionally, parsing the response with BeautifulSoup
soup = BeautifulSoup(response.text, 'html.parser')

# For example, printing the page title or any other element
page_title = soup.title.text
print(f"Page title: {page_title}")
