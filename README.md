# Exam Bot for Discord
A simple Discord bot written in Go to help manage and track exam dates within a Discord server.

## Features
- Add new exams with a name and date (`!addexam <exam_name> <dd.mm.yyyy>`)
- Check time left until a specified exam (`!timeleft <exam_name>`)
- Update existing exam dates (`!updateexam <exam_name> <dd.mm.yyyy>`)
- Delete exams (`!deleteexam <exam_name>`)
- List all scheduled exams (`!listexams`)

## Commands
| Command       | Description                     | Usage                         |
|---------------|---------------------------------|-------------------------------|
| `!addexam`    | Add a new exam with a date      | `!addexam Math 30.06.2025`    |
| `!timeleft`   | Show time left until the exam   | `!timeleft Math`              |
| `!updateexam` | Update date of an existing exam | `!updateexam Math 01.07.2025` |
| `!deleteexam` | Remove an exam                  | `!deleteexam Math`            |
| `!listexams`  | List all exams                  | `!listexams`                  |

## Date Format
All exam dates must be provided in the `dd.mm.yyyy` format.

## Notes
- The bot prevents adding or updating exams to past dates.
- It checks for duplicate exam names.
- If an exam does not exist for certain commands, the bot informs the user.
