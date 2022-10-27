Name:           tanzu-cli
Version:        0.26.0
Release:        1%{?dist}
Summary:        The Tanzu CLI

%ifarch amd64
BuildArch:      x86_64
%endif

%ifarch arm64
BuildArch:      aarch64
%endif

License:        Apache 2.0
Source0:        tanzu-cli.tar.gz

%description
Install the Tanzu CLI

%prep
%setup -q

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/%{_bindir}
cp tanzu-core-linux_amd64 $RPM_BUILD_ROOT/%{_bindir}/tanzu

%clean
rm -rf $RPM_BUILD_ROOT

%files
%{_bindir}/tanzu
