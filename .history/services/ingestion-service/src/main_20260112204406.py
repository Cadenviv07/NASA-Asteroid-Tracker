import requests
from datetime import date
from typing import TypedDict

API_KEY = 
today = date.today().strftime("%Y-%m-%d")
def getTodaysMeteors():
    url = https://api.nasa.gov/neo/rest/v1/feed?start_date={today}&end_date={today}&{api_key}=API_KEY
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
                asteroid_ai = a["id"]

                details_url = f"https://api.nasa.gov/neo/rest/v1/neo/{asteroid_id}?api_key={API_KEY}"
                
                details = requests.get(details_url).json()

                rock: AsteroidPayload = {
                    id: asteroid_ai,
                    asteroid: a["name"],
                    diameter_km: a["estimated_diameter"]["kilometers"]["estimated_diameter_max"],
                    velocity_kph: float(a["close_approach_data"][0]["relative_velocity"]["kilometers_per_hour"]),
                    orbital_elements: details["orbital_data"]
                }
                rocks.append(rock)
                print(f"Processed: {rock['asteroid']}")
            return rocks             
        else:
            print(f"Error: {response.status_code}")
    except requests.exceptions.RequestException as e:
        print(f"An error occured: {e}")

class AsteroidPayload(TypedDict):
    id: str
    asteroid: str
    diameter_km: float
    velocity_kph: float
    orbital_elements:Dict