# coding=utf-8
from setuptools import setup

from os.path import join, dirname

setup(
    name='ipgeobase-importer',
    version='1.6.2',
    packages=[],
    url='https://github.com/m-messiah/ipgeobase-importer',
    license='MIT',
    author='m_messiah',
    author_email='m.muzafarov@gmail.com',
    description=u'Импорт ipgeobase, maxmind (py3 only) и TOR баз '
                u'в совместимые с nginx geoIP map-файлы',
    long_description=open(join(dirname(__file__), 'README.rst')).read(),
    scripts=['ipgeobase-importer', 'ip-maxmind'],
    install_requires=['requests', 'iptools'],
    keywords='ipgeobase tor nginx geoip',
    classifiers=[
        'Development Status :: 5 - Production/Stable',
        'Intended Audience :: System Administrators',
        'Topic :: Internet :: WWW/HTTP',
        'Topic :: Utilities',
        'License :: OSI Approved :: MIT License',
        'Programming Language :: Python :: 2',
        'Programming Language :: Python :: 2.7',
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.3',
        'Programming Language :: Python :: 3.4',
        'Programming Language :: Python :: 3.5',
    ],
)
