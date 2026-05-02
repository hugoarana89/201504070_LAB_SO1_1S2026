from locust import HttpUser, task, between
import random
from datetime import datetime, timedelta

class MilitaryReportUser(HttpUser):
    wait_time = between(0.1, 0.5)

    def random_timestamp_last_year(self):
        now = datetime.utcnow()
        one_year_ago = now - timedelta(days=365)

        # Diferencia total en segundos
        delta_seconds = int((now - one_year_ago).total_seconds())

        # Elegir un segundo aleatorio dentro del rango
        random_seconds = random.randint(0, delta_seconds)

        random_date = one_year_ago + timedelta(seconds=random_seconds)

        # Formato ISO 8601 con Z (UTC)
        return random_date.strftime("%Y-%m-%dT%H:%M:%SZ")

    @task
    def send_report(self):
        countries = ["USA", "RUS", "CHN", "ESP", "GMT"]

        payload = {
            "country": random.choice(countries),
            "warplanes_in_air": random.randint(0, 50),
            "warships_in_water": random.randint(0, 30),
            "timestamp": self.random_timestamp_last_year()
        }

        self.client.post("/grpc-201504070", json=payload)