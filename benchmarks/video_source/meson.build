project('testproject', 'c',
  version : '0.1',
  default_options : ['warning_level=3'])

serverdebs = [dependency('gstreamer-1.0'), dependency('gstreamer-rtsp-server-1.0')]

executable('server',
           'server.c', dependencies: serverdebs)
