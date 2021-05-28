#include <stdio.h>
#include <limits.h>
#include <time.h>
#include <errno.h>
#include <stdlib.h>
#include <string.h>

#include <hs/hs.h>

struct list
{
    char *pattern;
    int index;
    int length;
    struct list *next;
};

typedef struct list LIST;

char *getLineOfAnySize(FILE *fp, int *endOfLineDetected, int *length)
{
    char *line;            // buffer for our string
    int ch;                // we will read line character by character
    int len = 0;           // number of characters read (character counter)
    size_t lineSize = 128; // initial size of the buffer allocated for the line
    *length = 0;

    if (!fp)
        return NULL;

    // allocating the buffer
    line = (char *)realloc(NULL, sizeof(char) * lineSize); // expected size of the line is up to typicalSize

    if (!line)
        return line; // protection, if we fail to allocate the memory we will return NULL

    while (1)
    {
        ch = fgetc(fp);

        if (ch == '\n')
            break;
        if (ch == EOF)
        {
            *endOfLineDetected = 1;
            break;
        }

        line[len++] = ch;

        if (len == lineSize)
        {
            lineSize = lineSize + 64;
            line = (char *)realloc(line, sizeof(char) * (lineSize));
            if (!line)
                return line;
        }
    }

    if ((len == 0) && *endOfLineDetected)
        return NULL;

    line[len++] = '\0';
    *length = len;

    return line;
}

void readRegex(const char *inputFN, LIST **head)
{
    FILE *fp = fopen(inputFN, "rb");
    if (!fp)
    {
        fprintf(stderr, "ERROR: unable to open file \"%s\": %s\n", inputFN,
                strerror(errno));
        return;
    }

    char *line;
    int endOfLineDetected = 0;
    int nrOfCharRead = 0;
    int index = 0;

    LIST *current;
    *head = current = NULL;

    while (line = getLineOfAnySize(fp, &endOfLineDetected, &nrOfCharRead))
    {
        if ((nrOfCharRead == 0) && endOfLineDetected)
            break;

        LIST *node = (LIST *)malloc(sizeof(LIST));

        node->pattern = line;
        node->index = index;
        node->next = NULL;
        node->length = nrOfCharRead;

        if (*head == NULL)
            current = *head = node;
        else
            current = current->next = node;

        if (endOfLineDetected)
            break;
        nrOfCharRead = 0;
        index++;
    }

    fclose(fp);
}

static char *readInputData(const char *inputFN, unsigned int *length)
{
    FILE *f = fopen(inputFN, "rb");
    if (!f)
    {
        fprintf(stderr, "ERROR: unable to open file \"%s\": %s\n", inputFN,
                strerror(errno));
        return NULL;
    }

    /* We use fseek/ftell to get our data length, in order to keep this example
     * code as portable as possible. */
    if (fseek(f, 0, SEEK_END) != 0)
    {
        fprintf(stderr, "ERROR: unable to seek file \"%s\": %s\n", inputFN,
                strerror(errno));
        fclose(f);
        return NULL;
    }
    long dataLen = ftell(f);
    if (dataLen < 0)
    {
        fprintf(stderr, "ERROR: ftell() failed: %s\n", strerror(errno));
        fclose(f);
        return NULL;
    }
    if (fseek(f, 0, SEEK_SET) != 0)
    {
        fprintf(stderr, "ERROR: unable to seek file \"%s\": %s\n", inputFN,
                strerror(errno));
        fclose(f);
        return NULL;
    }

    /* Hyperscan's hs_scan function accepts length as an unsigned int, so we
     * limit the size of our buffer appropriately. */
    if ((unsigned long)dataLen > UINT_MAX)
    {
        dataLen = UINT_MAX;
        printf("WARNING: clipping data to %ld bytes\n", dataLen);
    }
    else if (dataLen == 0)
    {
        fprintf(stderr, "ERROR: input file \"%s\" is empty\n", inputFN);
        fclose(f);
        return NULL;
    }

    char *inputData = (char *)malloc(dataLen);
    if (!inputData)
    {
        fprintf(stderr, "ERROR: unable to malloc %ld bytes\n", dataLen);
        fclose(f);
        return NULL;
    }

    char *p = inputData;
    size_t bytesLeft = dataLen;
    while (bytesLeft)
    {
        size_t bytesRead = fread(p, 1, bytesLeft, f);
        bytesLeft -= bytesRead;
        p += bytesRead;
        if (ferror(f) != 0)
        {
            fprintf(stderr, "ERROR: fread() failed\n");
            free(inputData);
            fclose(f);
            return NULL;
        }
    }

    fclose(f);

    *length = (unsigned int)dataLen;
    return inputData;
}

static int eventHandler(unsigned int id, unsigned long long from,
                        unsigned long long to, unsigned int flags, void *ctx)
{
    //printf("Match from %llu to %llu\n", from, to);
    return 0;
}

static long match(char *regex, char *input, int length)
{
    // compile regex
    hs_database_t *database;
    hs_compile_error_t *compile_err;

    if (hs_compile(regex, HS_FLAG_ALLOWEMPTY, HS_MODE_BLOCK, NULL, &database,
                   &compile_err) != HS_SUCCESS)
    {
        fprintf(stderr, "ERROR: Unable to compile pattern \"%s\": %s\n",
                regex, compile_err->message);
        hs_free_compile_error(compile_err);
        return -1;
    }

    // create scratch
    hs_scratch_t *scratch = NULL;
    if (hs_alloc_scratch(database, &scratch) != HS_SUCCESS)
    {
        fprintf(stderr, "ERROR: Unable to allocate scratch space. Exiting.\n");
        hs_free_database(database);
        return -1;
    }

    // match
    struct timespec tp1, tp2;
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &tp1);
    if (hs_scan(database, input, length, 0, scratch, eventHandler, NULL) != HS_SUCCESS)
    {
        fprintf(stderr, "ERROR: Unable to scan input buffer. Exiting.\n");
        hs_free_scratch(scratch);
        hs_free_database(database);
        return -1;
    }
    clock_gettime(CLOCK_PROCESS_CPUTIME_ID, &tp2);
    long t_diff = ((tp2.tv_sec - tp1.tv_sec) * (1000 * 1000 * 1000) + (tp2.tv_nsec - tp1.tv_nsec)) / 1000;

    hs_free_scratch(scratch);
    hs_free_database(database);
    return t_diff;
}

int main(int argc, char *argv[])
{
    if (argc != 3)
    {
        fprintf(stderr, "Usage: %s <input file> <regex file>\n", argv[0]);
        return -1;
    }
    char *inputFN = argv[1];
    char *regexFN = argv[2];

    LIST *regex_head;
    readRegex(regexFN, &regex_head); // read the regex expression into list

    LIST *str_head;
    readRegex(inputFN, &str_head);

    // unsigned int length = 0;
    // char *inputData = readInputData(inputFN, &length);

    long time_spent = 0;

    LIST *curr_regex = regex_head;
    LIST *curr_str = str_head;

    while (curr_regex && curr_str)
    {
        time_spent = 0;
        for (int i = 0; i < 1; i++)
        {
            time_spent += match(curr_regex->pattern, curr_str->pattern, curr_str->length);
        }
        printf("%lu\n", time_spent);
        curr_regex = curr_regex->next;
        curr_str = curr_str->next;
    }
    return 0;
}