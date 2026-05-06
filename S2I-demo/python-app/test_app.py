import pytest
from app import app

@pytest.fixture
def client():
    with app.test_client() as client:
        yield client

def test_hello_page(client):
    """Anasayfanın 200 dönüp dönmediğini ve içeriği kontrol eder."""
    rv = client.get('/')
    assert rv.status_code == 200
    assert b"Hello from Python S2I!" in rv.data
