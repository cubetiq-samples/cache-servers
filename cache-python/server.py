from flask import Flask, request
from cache import cache

app = Flask(__name__)


@app.route('/cache', methods=['GET', 'POST'])
def cache_api():
    if request.method == 'GET':
        # Retrieve value from cache
        key = request.args.get('key')
        if key in cache:
            response = {'key': key, 'value': cache[key]}
        else:
            response = {'error': 'Key not found'}

    elif request.method == 'POST':
        # Store value in cache
        key = request.form.get('key')
        value = request.form.get('value')
        cache[key] = value
        response = {'message': 'Value stored'}

    else:
        response = {'error': 'Invalid request method'}

    return response


if __name__ == '__main__':
    app.run()
