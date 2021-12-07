Name:          dynflare
Version:       1.0.0
Release:       1
Summary:       Dynamic DNS updater without polling external services.
License:       BSD-3-Clause
URL:           https://github.com/lukasdietrich/dynflare
Packager:      Lukas Dietrich <lukas@lukasdietrich.com>
BuildArch:     x86_64
Requires:      systemd
Source0:       dynflare
Source1:       dynflare.service
Source2:       example.config.toml

%description
%{summary}

%pre
getent group dynflare >/dev/null || groupadd -r dynflare
getent passwd dynflare >/dev/null || useradd -r -M -s /bin/false -c "Dynflare Daemon" -g dynflare dynflare

%install
install -p -D -m 755 %{SOURCE1} %{buildroot}%{_bindir}/dynflare
install -p -D -m 644 %{SOURCE1} %{buildroot}%{_unitdir}/dynflare.service
install -p -D -m 600 %{SOURCE2} %{buildroot}%{_sysconfdir}/dynflare/config.toml
install -d -m 755 %{buildroot}%{_localstatedir}/cache/dynflare

%files
%attr(755,root,root) %{_bindir}/dynflare
%attr(644,root,root) %{_unitdir}/dynflare.service
%config(noreplace) %attr(600,dynflare,dynflare) %{_sysconfdir}/dynflare/config.toml
%dir %attr(700,dynflare,dynflare) %{_localstatedir}/cache/dynflare

%post
%systemd_post dynflare.service

%preun
%systemd_preun dynflare.service

%postun
%systemd_poston_with_restart dynflare.service
