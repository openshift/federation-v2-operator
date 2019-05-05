#!/bin/bash

from_version=$1
to_version=$2

dir=$(realpath "$(dirname "${BASH_SOURCE}")/..")

declare -a packages=("federation" "cluster-federation")
for package in "${packages[@]}"
do
  echo "working on package $package"
  package_dir=$dir/upstream-manifests/$package
  from_version_dir=$package_dir/$from_version
  to_version_dir=$package_dir/$to_version
  
  git mv $from_version_dir/$package.v$from_version.clusterserviceversion.yaml $from_version_dir/$package.v$to_version.clusterserviceversion.yaml
  git mv $from_version_dir $to_version_dir
  find $to_version_dir -type f -name *.yaml -exec sed -i -e "s,$from_version,$to_version,g" '{}' \;
  sed "s,$from_version,$to_version,g" -i $package_dir/$package.package.yaml
done

declare -a scrub_dirs=("upstream-manifests/$package/$to_version" "olm-testing" "scripts")
for scrub_dir in "${scrub_dirs[@]}"
do
  find $dir/$scrub_dir -type f -exec sed -i -e "s,$from_version,$to_version,g" '{}' \;
done

declare scrub_files=("Makefile.ci" "Dockerfile" "Gopkg.toml")
for scrub_file in "${scrub_files[@]}"
do
  sed "s,$from_version,$to_version,g" -i $dir/$scrub_file
done

