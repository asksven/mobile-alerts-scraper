#!/bin/bash

docker build -t asksven/mobile-alerts-scraper:latest .
docker push asksven/mobile-alerts-scraper:latest

docker build -t asksven/mobile-alerts-scraper:raspi-latest -f $(pwd)/Dockerfile.raspi .
docker push asksven/mobile-alerts-scraper:raspi-latest
