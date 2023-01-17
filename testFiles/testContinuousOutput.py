import time
import logging
import random

log_output = [
    "Maecenas placerat turpis diam, ut venenatis tellus commodo non. Ut vel sapien ut lorem pharetra feugiat non quis lorem.",
    "Integer varius ipsum lobortis dolor faucibus, in malesuada lectus venenatis. Pellentesque vulputate orci scelerisque tincidunt convallis.",
    "Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. In elementum tempor magna eget tincidunt.",
    "Ut vel sapien ut lorem pharetra feugiat non quis lorem. Sed volutpat erat ut sapien venenatis sollicitudin. Pellentesque et rutrum eros.",
]

output_method = [
    logging.info,
    logging.error,
    logging.warning,
    logging.critical
]

logging.basicConfig()
logging.root.setLevel(logging.DEBUG)

while True:
    message = log_output[random.randrange(0, len(log_output))]
    method = output_method[random.randrange(0, len(output_method))]

    method(message)
    time.sleep(random.randrange(1, 3))

