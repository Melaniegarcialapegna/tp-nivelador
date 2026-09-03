import sys 

#constants
OUTPUT_FILE = "docker-compose.yaml"

SERVER_CONTENT = """services:
  server:
    build:
      context: ./services/server
      dockerfile: Dockerfile
    container_name: server
    ports:
      - "5678:5678"
    environment:
      - PYTHONUNBUFFERED=1
      - SERVER_HOST=server
      - SERVER_PORT=5678
      - SERVER_STORAGE_PATH=/output/bets_storage.csv
      - AGENCY_QUORUM_MIN={i} 
    volumes:
      - ./output:/output
"""
 
CLIENT_CONTENT = """
  client_{i}:
    build:
      context: ./services/client
      dockerfile: Dockerfile
    container_name: client_{i}
    depends_on:
      - server
    environment:
      - AGENCY_ID={i}
      - SERVER_HOST=server
      - SERVER_PORT=5678
      - INPUT_FILE=/input/input-{i}.csv
      - OUTPUT_FILE=/output/output-{i}.csv
      - BATCH_SIZE={j}
    volumes:
      - ./input:/input
      - ./output:/output
"""

def main():
    number_of_clients = parse_number_of_clients()
    docker_compose_content = generate_docker_compose(number_of_clients)

    #Write the content in the file
    #In case the file does not exit it is created
    with open(OUTPUT_FILE, "w") as file:
        file.write(docker_compose_content)

    print(f"The file {OUTPUT_FILE} was successfully generated with {number_of_clients} client/s")

    
def parse_number_of_clients():
    if len(sys.argv)!=2: 
        print("Error: invalid number of arguments")
        print("Valid command: python3 generator_docker-compose.py <number_of_clients>")
        sys.exit(1)

    try:
        number_of_clients = int(sys.argv[1])
    except ValueError:
        print("Error: the number of clients has to be an integer")
        sys.exit(1)

    if number_of_clients < 1:
        print("Error: the number of clients has to be greater than 0")
        sys.exit(1)

    return number_of_clients

def generate_docker_compose(number_of_clients, batch_size=5, quorum_min=5):
    file_content = ""
    file_content += SERVER_CONTENT.format(i=quorum_min)
    for i in range (number_of_clients):
        file_content+=CLIENT_CONTENT.format(i=i,j=batch_size)

    file_content = file_content.rstrip()

    return file_content

if __name__ == "__main__":
    main()