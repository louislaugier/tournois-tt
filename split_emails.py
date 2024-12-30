#!/usr/bin/env python3

def split_csv():
    input_file = '~/Documents/clubs.csv'
    output_file1 = 'first_2000_emails.csv'
    output_file2 = 'remaining_emails.csv'
    
    # Read all emails, removing empty lines and stripping whitespace
    with open(input_file.replace('~', os.path.expanduser('~')), 'r') as f:
        emails = [line.strip() for line in f.readlines() if line.strip()]
    
    # Remove header if it exists
    if emails and emails[0].lower() == 'email;':
        emails = emails[1:]
    
    # Remove trailing semicolons
    emails = [email.rstrip(';') for email in emails]
    
    # Split into two parts
    first_2000 = emails[:2000]
    remaining = emails[2000:]
    
    # Write first 2000 emails
    with open(output_file1, 'w') as f:
        f.write('email\n')
        for email in first_2000:
            f.write(f'{email}\n')
    
    # Write remaining emails
    with open(output_file2, 'w') as f:
        f.write('email\n')
        for email in remaining:
            f.write(f'{email}\n')
    
    print(f'Split complete!')
    print(f'First file ({output_file1}): {len(first_2000)} emails')
    print(f'Second file ({output_file2}): {len(remaining)} emails')

if __name__ == '__main__':
    import os
    split_csv() 