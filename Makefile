


build:
	@echo "Fake building"

canary: build
	@echo "Fake canary building"

stable: build
	@echo "fake stable"

rollback:
	@echo "Rolling back canary"