import requests
from datetime import date
from typing import TypedDict, Dict
import os
from dotenv import load_dotenv
import boto3
import json

load_dotenv()

API_KEY = os.getenv("NASA_API_KEY")

if not API_KEY:
    raise ValueError('NO API KEY FOUND')
else:
    print(f"2. API Key found: {API_KEY[:4]}******")

sqs = boto3.client('sqs', region_name='us-east-2')

QueueUrl = "https://sqs.us-east-2.amazonaws.com/574070665369/asteroidBelt"


class AsteroidPayload(TypedDict):
    id: str
    asteroid: str
    diameter_km: float
    velocity_kph: float
    orbital_elements:Dict

today = date.today().strftime("%Y-%m-%d")
def getTodaysMeteors():
    #First api gets a list of all the different meteorite ids of today
    url = f"https://api.nasa.gov/neo/rest/v1/feed?start_date={today}&end_date={today}&api_key={API_KEY}"
    response =   requests.get(url)
    try:
        if response.status_code == 200:
            data = response.json()

            if data["element_count"] == 0:
                print("No asteroids found.")
                return

            asteroids = data["near_earth_objects"][today]
            rocks = []
            for a in asteroids:
                asteroid_id = a["id"]
                #Second api calls every specific meteorite to get information on them
                details_url = f"https://api.nasa.gov/neo/rest/v1/neo/{asteroid_id}?api_key={API_KEY}"
                
                details = requests.get(details_url).json()

                rock: AsteroidPayload = {
                    "id": asteroid_id,
                    "asteroid": a["name"],
                    "diameter_km": a["estimated_diameter"]["kilometers"]["estimated_diameter_max"],
                    "velocity_kph": float(a["close_approach_data"][0]["relative_velocity"]["kilometers_per_hour"]),
                    "orbital_elements": details["orbital_data"]
                }
                rocks.append(rock)
                
                message_body = json.dumps(rock)
                
                sqs.send_message(
                    QueueUrl = QueueUrl,
                    MessageBody = message_body
                )

                print(f"Processed: {rock['asteroid']}")
            return rocks             
        else:
            print(f"Error: {response.status_code}")
    except requests.exceptions.RequestException as e:
        print(f"An error occured: {e}")

if __name__ == "__main__":
    getTodaysMeteors()

