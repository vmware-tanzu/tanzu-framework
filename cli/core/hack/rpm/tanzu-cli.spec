Name:           tanzu-cli
Version:        0.26.0
Release:        1%{?dist}
Summary:        The Tanzu CLI
BuildArch:      x86_64

License:        Apache 2.0
Source0:        %{name}-linux-%{arch}.tar.gz


%description
The tanzu CLI

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

