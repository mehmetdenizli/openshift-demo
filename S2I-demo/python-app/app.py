from flask import Flask
import os
import socket

app = Flask(__name__)

@app.route('/')
def hello():
    html = "<h3>Hello from Python S2I!</h3>" \
           "<b>Hostname:</b> {hostname}<br/>" \
           "<b>App Version:</b> {version}<br/>"
    return html.format(hostname=socket.gethostname(), version=os.getenv("APP_VERSION", "1.0.0"))

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=8080)
