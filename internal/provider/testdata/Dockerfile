# KDC container
FROM centos/systemd:latest as kdc

EXPOSE 88
EXPOSE 464

STOPSIGNAL SIGRTMIN+3

# Install Kerberos client packages and copy in configuration
RUN yum install -y krb5-workstation && yum update -y && yum clean all
COPY --chown=root:root krb5.conf /etc/krb5.conf
RUN chmod 644 /etc/krb5.conf

# Install Kerberos server packages and copy in configuration
RUN yum install -y krb5-server && yum clean all
COPY --chown=root:root kdc.conf /var/kerberos/krb5kdc/kdc.conf
COPY --chown=root:root kadm5.acl /var/kerberos/krb5kdc/kadm5.acl
RUN chmod 600 /var/kerberos/krb5kdc/kdc.conf /var/kerberos/krb5kdc/kadm5.acl

# Enable Kerberos services
RUN systemctl enable krb5kdc.service kadmin.service

# Create Kerberos realm with random password
RUN kdb5_util create -s -r EXAMPLE.COM -P $(echo ${RANDOM}${RANDOM}${RANDOM} | md5sum | cut -d ' ' -f 1)

# Create test@EXAMPLE.COM principal with a password of 'password' and write
# it out to a keytab
RUN kadmin.local addprinc -pw password test
RUN kadmin.local ktadd -norandkey -k /etc/test.keytab test

# Create DNS/ns.example.com principal for BIND and write it out to a keytab
RUN kadmin.local addprinc -randkey DNS/ns.example.com
RUN kadmin.local ktadd -k /etc/named.keytab DNS/ns.example.com

# BIND container
FROM centos/systemd:latest as ns

EXPOSE 53

STOPSIGNAL SIGRTMIN+3

# Systemd backport
RUN yum -y install  wget
RUN wget --no-check-certificate https://copr.fedorainfracloud.org/coprs/jsynacek/systemd-backports-for-centos-7/repo/epel-7/jsynacek-systemd-backports-for-centos-7-epel-7.repo -O /etc/yum.repos.d/jsynacek-systemd-centos-7.repo
RUN yum -y update systemd ; yum clean all

# Install Kerberos client packages (should use cached KDC step above)
RUN yum install -y krb5-workstation && yum update -y && yum clean all
COPY --chown=root:root krb5.conf /etc/krb5.conf
RUN chmod 644 /etc/krb5.conf

# Install BIND packages and copy in the keytab created by the KDC
RUN yum install -y bind bind-utils && yum clean all
COPY --from=kdc --chown=root:named /etc/named.keytab /etc/named.keytab
RUN chmod 640 /etc/named.keytab

# Enable BIND service
RUN systemctl enable named.service

# Copy in default zone content
COPY --chown=named:named db.* /var/named/dynamic/
RUN chmod 644 /var/named/dynamic/db.*

# Keytab container
FROM scratch as keytab

# Copy in the keytab created by the KDC
COPY --from=kdc /etc/test.keytab /test.keytab
