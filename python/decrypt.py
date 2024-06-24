from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.backends import default_backend
import os

def decrypt_file(key, input_filename, output_filename):
    # Read the encrypted file
    with open(input_filename, 'rb') as f:
        nonce = f.read(12)  # Read the nonce (12 bytes for GCM)
        ciphertext = f.read()  # Read the rest of the file

    # Separate the ciphertext and the tag
    tag = ciphertext[-16:]  # The last 16 bytes are the tag
    ciphertext = ciphertext[:-16]  # The rest is the actual ciphertext

    # Create a Cipher object to perform the decryption
    cipher = Cipher(algorithms.AES(key), modes.GCM(nonce, tag), backend=default_backend())
    decryptor = cipher.decryptor()

    # Decrypt the data
    decrypted_data = decryptor.update(ciphertext) + decryptor.finalize()

    # Write the decrypted data to the output file
    with open(output_filename, 'wb') as f:
        f.write(decrypted_data)

if __name__ == "__main__":
    key = input("Enter the 32-byte key: ").encode()  # Prompt for the key
    input_filename = input("Enter the path to the encrypted file: ")  # Prompt for the input file path
    output_filename = input("Enter the path for the decrypted output file: ")  # Prompt for the output file path

    decrypt_file(key, input_filename, output_filename)
    print(f"Decryption complete. Decrypted file saved as {output_filename}")
