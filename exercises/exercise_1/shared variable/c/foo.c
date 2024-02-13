// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly

#include <pthread.h>
#include <stdio.h>

int i = 0;

//// Choose mutex because we want a binary lock that is either locked or unlocked. 
//// Semaphores are for synchronizeing variable access for multiuse,
//// for example: We want only two threads can access a resource any given time. 
pthread_mutex_t mutex;

// Note the return type: void*
void* incrementingThreadFunction(){
    // TODO: increment i 1_000_000 times
    for (int a = 0; a < 1000000; a++){
        pthread_mutex_lock(&mutex);
        i++;
        pthread_mutex_unlock(&mutex);
    }
    return NULL;
}

void* decrementingThreadFunction(){
    // TODO: decrement i 1_000_000 times
    for (int a = 0; a < 999999; a++){
        pthread_mutex_lock(&mutex);
        i--;
        pthread_mutex_unlock(&mutex);
    }
    
    return NULL;
}


int main(){
    // TODO: 
    // start the two functions as their own threads using `pthread_create`
    // Hint: search the web! Maybe try "pthread_create example"?

    pthread_t incThread;
    pthread_t decThread;

    pthread_mutex_init(&mutex , NULL);


    int incStat;
    int decStat;

    incStat = pthread_create(&incThread, NULL, incrementingThreadFunction, NULL );
    // printf("increment status: %d\n", incStat);

    decStat = pthread_create(&decThread, NULL, decrementingThreadFunction, NULL);
    // printf("decrement status: %d\n", decStat);


    // TODO:
    // wait for the two threads to be done before printing the final result
    // Hint: Use `pthread_join`    
    pthread_join(incThread, NULL);
    pthread_join(decThread, NULL);

    
    printf("The magic number is: %d\n", i);
    return 0;
}
