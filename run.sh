go build -o bed-and-breakfast cmd/web/*.go
./bed-and-breakfast -dbname=bookings -dbuser=postgres -dbpassword=password -cache=false -production=false