# Results

## 3: Sharing a variable
In both the c code and the go code the resulting number is unpredictable. This is because both the increment thread and the decrement thread shares the same resource and a race condition occurs. When the scheduler changes which task that gets to run on the CPU, the state of the task that was running i is saved on the stack, and the state of the new thread are loaded in. If this for example happens mid incrementing, then the variable will be loaded, at lets say 5. Then the decrement thread take over, and load in the variable as 5 and decrement it to 4. When the increment thread then gets time to run it will load in from the saved state, it will therefore load in the variable as 5 and increment it to 6. The decrementation operation was now "erased", since we have no control over when the scheduler does this context switching the variable is not actually incremented and decremented 1000000 times but some of the operations disappear and we get an unpredictable result. 



