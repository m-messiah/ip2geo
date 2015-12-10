# coding=utf-8
from distutils.core import setup

from os.path import join, dirname

setup(
    name='ipgeobase-importer',
    version='1.2',
    packages=[''],
    url='https://github.com/m-messiah/ipgeobase-importer',
    license='MIT',
    author='m_messiah',
    author_email='m.muzafarov@gmail.com',
    description=u'Импорт ipgeobase и TOR баз '
                u'в совместимые с nginx geoIP map-файлы',
    long_description=open(join(dirname(__file__), 'README.rst')).read(),
    scripts=['ipgeobase-importer'],
    install_requires=['requests', ],
    keywords='ipgeobase tor nginx geoip',
    classifiers=[
        # How mature is this project? Common values are
        #   3 - Alpha
        #   4 - Beta
        #   5 - Production/Stable
        'Development Status :: 5 - Production/Stable',

        # Indicate who your project is intended for
        'Intended Audience :: System Administrators',
        'Topic :: Internet :: WWW/HTTP',
        'Topic :: Utilities',

        # Pick your license as you wish (should match "license" above)
        'License :: OSI Approved :: MIT License',

        # Specify the Python versions you support here. In particular, ensure
        # that you indicate whether you support Python 2, Python 3 or both.
        'Programming Language :: Python :: 2',
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.2',
        'Programming Language :: Python :: 3.3',
        'Programming Language :: Python :: 3.4',
        'Programming Language :: Python :: 3.5',
    ],
)
