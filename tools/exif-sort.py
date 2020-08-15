#!/usr/bin/python3

import sys
import datetime
import exiftool
import os
import shutil
import logging


def get_key(path):
    parts = path.split(".")
    return parts[0]


def get_destination(group):
    with exiftool.ExifTool() as extool:
        for path in group:
            if path.endswith(".jpg"):
                meta = extool.get_metadata(path)
                ts_string = meta["File:FileAccessDate"]
                print(ts_string)
                if ts_string:
                    time, tz = ts_string.split("-")
                    print(time, tz)
                    ts = datetime.datetime.strptime(time, "%Y:%m:%d %H:%M:%S")
                    return ts.strftime("%Y%m/%d")
                else:
                    logging.info("%s: no exif", path)
    return None


def main():
    logging.basicConfig(stream=sys.stdout, level=logging.INFO)
    log = logging.getLogger("exifsort")

    start = "./"

    groups = {}
    files = [f for f in os.listdir(start) if os.path.isfile(os.path.join(start, f))]
    for fn in files:
        key = get_key(fn)
        path = os.path.join(start, fn)
        groups.setdefault(key, []).append(path)

    for key in groups:
        group = groups[key]
        new_path = get_destination(group)
        if new_path:
            logging.info("moving %s group to %s", key, new_path)
            os.makedirs(new_path, exist_ok=True)
            for child in group:
                child_name = os.path.basename(child)
                shutil.move(child, os.path.join(new_path, child_name))


if __name__ == "__main__":
    main()
