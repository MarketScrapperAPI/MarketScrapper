FROM golang

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/MarketScrapperAPI/MarketScrapper

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

ARG target=Auchan

# Run the executable
CMD ["MarketScrapper -t ${target}"]