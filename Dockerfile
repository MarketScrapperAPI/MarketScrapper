FROM golang

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/MarketScrapperAPI/MarketScrapper

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# Run the executable (being triggered by k8s)
# CMD ["MarketScrapper","-t","Auchan"]