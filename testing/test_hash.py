from hashlib import sha1
from urllib.parse import urlencode

test_hash = sha1(b"sha1 this string").digest()
d = dict()
d["info"] = test_hash
print(urlencode(d))
