import os
from dotenv import load_dotenv
import requests
from requests.auth import HTTPBasicAuth

# Load variables from .env file
load_dotenv()

account_sid = os.getenv("ACCOUNT_SID")
auth_token = os.getenv("AUTH_TOKEN")
from_number = os.getenv("FROM_NUMBER")
to_number = os.getenv("TO_NUMBER")

twiml = "<Response><Say>This is a test call from your matcha stock bot.</Say></Response>"

response = requests.post(
    f"https://api.twilio.com/2010-04-01/Accounts/{account_sid}/Calls.json",
    data={
        "To": to_number,
        "From": from_number,
        "Twiml": twiml,
    },
    auth=HTTPBasicAuth(account_sid, auth_token),
)

print(response.status_code)
print(response.json())
