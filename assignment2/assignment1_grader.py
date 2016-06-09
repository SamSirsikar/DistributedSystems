"""
Script to test assignment 1. Subject to change.
"""
try:
    import requests
    import datetime
    import json
except Exception as e:
    print "Requests library not found. Please install it. \nHint: pip install requests"

person = {
    "email": "foo@gmail.com",
    "zip": "95110",
    "country": "U.S.A",
    "profession": "student",
    "favorite_color": "blue",
    "is_smoking": "no",
    "favorite_sport": "hiking",
    "food": {
        "type": "vegetarian",
        "drink_alcohol": "yes"
    },
    "music": {
        "spotify_user_id": "wizzler"
    },
    "movie": {
        "tv_shows": ["The Big Bang Theory"],
        "movies": ["Taken"]
    },
    "travel": {
        "flight": {
            "seat": "aisle"
        }
    }
}
change_person = {
    "travel": {
        "flight": {
            "seat": "window"
        }
    },
    "favorite_sport": "football"
}


def test_post(url):
    post_url = "%s/profile" % url
    try:
        r = requests.post(post_url, data=json.dumps(person))
        if r.status_code == 201:
            d = datetime.datetime.now()
            print "POST Check successful. Time: %s" % d
        else:
            print "POST incorrect. Not working as expected."
    except requests.exceptions.ConnectionError:
        print "Server not running on %s" % url
        exit()


def test_get(url):
    get_url = "%s/profile/%s" % (url, person['email'])
    r = requests.get(get_url)
    if r.status_code == 200:
        d = r.json()
        try:
            d = d[0]
        except KeyError:
            pass
        if d["zip"] == person['zip']:
            if type(d['food']) is dict:
                if d['movie']['movies'][0] == person['movie']['movies'][0]:
                    t = datetime.datetime.now()
                    print "GET check successful. Time: %s" % t
    else:
        print "GET check failed"


def test_put(url):
    put_url = "%s/profile/%s" % (url, person['email'])
    r = requests.put(put_url, data=json.dumps(change_person))
    if r.status_code == 204:
        t = datetime.datetime.now()
        print "PUT request sent successfully. Time: %s" % t
    else:
        print "PUT failed."
    get_url = "%s/profile/%s" % (url, person['email'])
    r = requests.get(get_url)
    if r.status_code == 200:
        d = r.json()
        try:
            d = d[0]
        except KeyError:
            pass
        if d['travel']['flight']['seat'] == change_person['travel']['flight']['seat']:
            if d['favorite_sport'] == change_person['favorite_sport']:
                t = datetime.datetime.now()
                print "GET after Put successful. Time: %s" % t
    else:
        print "GET after PUT failed"


def test_delete(url):
    delete_url = "%s/profile/%s" % (url, person['email'])
    r = requests.delete(delete_url)
    if r.status_code == 204:
        print "DELETE status code check complete"
    else:
        print "DELETE failed."

    get_url = "%s/profile/%s" % (url, person['email'])
    r = requests.get(url)
    if r.status_code == 200:
        print "DELETE has not deleted the item."
    else:
        print "You have deleted the item"


url = "http://0.0.0.0:4001" # raw_input("Enter the url without trailing slash. Ex: http://localhost:3000:\n")
import time
test_post("http://0.0.0.0:4001")
#time.sleep(1)
test_get("http://0.0.0.0:4002")
#time.sleep(1)
test_put("http://0.0.0.0:4001")
#time.sleep(1)
test_delete("http://0.0.0.0:4002")