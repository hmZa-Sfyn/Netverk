from flask import Flask, jsonify, request, abort
import sys
import logging
import subprocess
from functools import wraps

app = Flask(__name__)

node_name = None
auth_username = None
auth_password = None
auth_token = None  # Variable to store the generated auth token
token_validity = 30  # Validity in seconds
shutdown_called = False  # Track if shutdown has been called

# Set up logging to output errors
logging.basicConfig(level=logging.DEBUG)

# Basic auth decorator
def check_auth(username, password):
    
    return username == auth_username and password == auth_password

def requires_auth(f):
    @wraps(f)
    def decorated(*args, **kwargs):
        auth = request.headers.get('Authorization')
        if auth != f"Bearer {auth_token}":
            logging.warning("Authorization failed!")
            return jsonify({'message': 'Authorization required!'}), 401
        return f(*args, **kwargs)
    return decorated

@app.route('/', methods=['GET'])
@app.route('/home', methods=['GET'])
@app.route('/home/html', methods=['GET'])
def index_html():
    routes = [rule.rule for rule in app.url_map.iter_rules()]
    routes_to_list = ""
    x = 0
    for i in routes:
        routes_to_list = routes_to_list + f"""<tr style="padding:10px;background:black;">
                                                <td style="padding:8px;color:lime;border: 3px solid red;">
                                                    <a>{x}</a>
                                                </td>
                                                <td style="padding:8px;color:lime;border: 3px solid red;">
                                                    <a href=\"{i}\">{i}</a>
                                                </td>
                                              </tr>"""
        x = x + 1
        
    return f"""
            <h1 style="font-family:'Courier New', Courier, monospace;font-size: 25px;color:lime;padding: 20px;margin: 10px;background:black;border-radius: 40px;">
                API Server for <u>{node_name}</u> is running!
                <br>
                You are now serving <u>{node_name}</u>, by <u>{auth_username}</u> running on <u>{request.host}</u>!
            </h2>
            <h3 style="font-family:'Courier New', Courier, monospace;font-size: 20px;color:lime;padding: 40px;margin: 10px;background:black;border-radius: 50px;">
                Routes Are
            </h3>
            
            <table class="table">
                <thead>
                    <tr>
                        <th>#</th>
                        <th>Link</th>
                    </tr>
                </thead>
                <tbody>
                    {routes_to_list}
                </tbody>
            </table>
"""

@app.route('/home/json', methods=['GET'])
def index_json():
    routes = [rule.rule for rule in app.url_map.iter_rules()]
    routes_to_list = ""
    for i in routes:
        routes_to_list = routes_to_list + f"<li>{i}</li>"
    return jsonify({'info': {"node":{"name":node_name,"owner":{f"{auth_username}":{"r":"1","w":"1","x":"1"},f"you":{"r":"1","w":"0","x":"0"}}},"host":request.host,"admin":{"username":auth_username}},"request":{"ip":request.host.split(":")[0],"port":request.host.split(":")[1]}}), 200

@app.route('/check/islive', methods=['GET'])
def check_islive():
    return jsonify({'info':"True"}), 200



@app.route('/hello', methods=['GET'])
def hello():
    return jsonify({'message': f'Greetings from {node_name}, running on {request.host}'}), 200

@app.route('/routes', methods=['GET'])
def list_routes():
    routes = [rule.rule for rule in app.url_map.iter_rules()]
    return jsonify({'routes': routes}), 200

@app.route('/robots.txt', methods=['GET'])
def list_robots_conditioned_routes():
    routes = {"Allowed":{"main":"./","db":"./db"},"Not Allowed":{"outer dir structure":"../"}}
    return jsonify({'routes': routes}), 200

@app.route('/f/auth/login', methods=['POST'])
def login():
    username = request.json.get('username')
    password = request.json.get('password')
    
    if check_auth(username, password):
        global auth_token
        auth_token = f"{username}-token-supersecret-please-dont-share-this-and-this-is-totally-very-secure!!!"  # Simple token generation
        logging.info("Login successful, token generated.")
        return jsonify({'auth-token': auth_token}), 200
    else:
        logging.warning("Login failed: invalid credentials.")
        return jsonify({'message': 'Invalid username or password!'}), 401

@app.route('/f/system/shutdown', methods=['POST'])
@requires_auth
def shutdown():
    shutdown_server()
    return jsonify({'message': 'Server is shutting down...'}), 200

def shutdown_server():
    func = request.environ.get('werkzeug.server.shutdown')
    if func is None:
        logging.error("Shutdown function is not available. Not running with the Werkzeug Server.")
        raise RuntimeError('Not running with the Werkzeug Server')
    func()
    logging.info("Server is shutting down.")


@app.route('/f/ssh/cmd', methods=['POST'])
@requires_auth
def run_command():
    command = request.json.get('command')
    if not command:
        return jsonify({'message': 'No command provided!'}), 400
    
    try:
        # Run the command and capture the output
        result = subprocess.run(command, shell=True, capture_output=True, text=True)
        response = {
            'stdout': result.stdout.strip(),
            'stderr': result.stderr.strip(),
            'return_code': result.returncode
        }
        return jsonify({'response': response}), 200
    except Exception as e:
        logging.error(f"Error running command: {e}")
        return jsonify({'message': 'Error running command!'}), 500

if __name__ == '__main__':
    if len(sys.argv) != 5:
        print("Usage: python basic_api_serv.py <node_name> <auth_username> <auth_password> <port>")
        sys.exit(1)

    node_name = sys.argv[1]
    auth_username = sys.argv[2]
    auth_password = sys.argv[3]
    port = int(sys.argv[4])

    logging.info(f"Starting server for node: {node_name} on port: {port}")
    app.run(host='0.0.0.0', port=port)
