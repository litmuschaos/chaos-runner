# Makefile for building Litmus and its tools
# Reference Guide - https://www.gnu.org/software/make/manual/make.html

#
# Internal variables or constants.
# NOTE - These will be executed when any make target is invoked.
#
.PHONY: chaos-executor
chaos-executor: 
	@echo "------------------"
	@echo "--> Build chaos-executor binary"
	go build bin/exec.go
	@echo "------------------"
	@echo "------------------"
	@echo "--> Build chaos-executor image" 
	@echo "------------------"
	sudo docker build . -f build/Dockerfile -t rahulchheda1997/chaos-executor:ci
	REPONAME="litmuschaos" IMGNAME="ansible-runner" IMGTAG="ci"