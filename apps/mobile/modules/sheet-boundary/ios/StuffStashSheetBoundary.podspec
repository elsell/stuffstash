Pod::Spec.new do |s|
  s.name = 'StuffStashSheetBoundary'
  s.version = '1.0.0'
  s.summary = 'Keyboard-window measurement for Stuff Stash native sheets'
  s.description = 'Converts the sheet boundary through its owning UIKit window.'
  s.homepage = 'https://github.com/elsell/stuffstash'
  s.license = { :type => 'UNLICENSED' }
  s.author = 'Stuff Stash contributors'
  s.source = { :git => 'https://github.com/elsell/stuffstash.git' }
  s.platforms = { :ios => '15.1' }
  s.swift_version = '5.9'
  s.static_framework = true
  s.dependency 'ExpoModulesCore'
  s.dependency 'React-Core'
  s.source_files = '**/*.{h,m,mm,swift}'
end
