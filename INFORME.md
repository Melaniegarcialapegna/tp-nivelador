# Informe TP Nivelador: Docker, Comunicaciones y Concurrencia

## Melanie Belén Garcia Lapegna ~ Padrón 111848

### Introducción
El trabajo presente esta compuesto por un servidor (en Python) que tiene el rol de simular una loteria, ademas se tienen N clientes (en Go), uno por cada agencia, que envian apuestas que tienen cargadas en su archivo `INPUT_FILE` y reciben como respuesta una lista con los ganadores pertenecientes a su agencia.
A continuacion se va a detallar el protocolo de comunicacion que se implemento y los mecanismos que se usaron para sincronizar la ejecución de manera concurrente.

### Capa de transporte (los `safe_sockets`)
Tanto en el cliente como en el servidor se implementaron las funciones `send_all`/`SendAll` y `recv_all`/`RecvAll` sobre los sockets TCP. 
Estas funciones evitan problemas de _short read_ y _short write_, lo que se hizo fue reenviar o reintentar la lectura en loop hasta que se completa la cantidad de bytes solicitada (en vez de asumir que un unico `send`/`recv` otorgan de una sola vez el mensaje completo como en la implementación propuesta en un principio).
Esta es **la unica capa que interactua de forma directa con el socket**.

### Protocolo de comunicación

#### Tipos de mensajes
Se definieron cuatro tipos de mensajes los cuales **se identifican por un byte al inicio de cada uno**:

| Tipo | Valor | Dirección | Finalidad |
|---|---|---|---|
| `BET` | 0 | ambas direcciones | avisa que lo proximo a leer es **un batch de apuestas**(si es del cliente al servidor) o avisa que se enviara **una apuesta ganadora** (si es del servidor al cliente) |
| `END` | 1 | ambas direcciones | avisa que termino de enviar los batchs de apuestas (en caso de que lo envie el cliente) o que termino de enviar los ganadores(servidor) |
| `ACK_OK` | 2 | de servidor a cliente | avisa al cliente que el batch fue procesado y almacenado correctamente |
| `ACK_FAIL` | 3 | de servidor a cliente | avisa al cliente que ocurrio un error al procesar el batch para que el mismo pueda tomar una decision al respecto (o no) |

#### Formato de los mensajes

Cada mensaje  de tipo `BET` o `END` se va leyendo de a 5 bytes para reducir lo maximo posible las llamadas a `recv` ya que implica un cambio te contexto costoso (desarrollado posteriormente).
En el caso de `BET` se compone de un **header fijo de 5 bytes**, el cual sigue de un **payload** que da idea de cual es la **longitud total del mensaje**:

```
[1 byte: tipo_de_mensaje] [4 bytes: longitud_del_payload] [payload...]
```
En este caso el payload es justamente el **batch** que es el conjunto de apuestas.
Entonces, basicamente el header lo que hace es permitir al receptor saber cuantos bytes tiene que leer a continuacion en el `recv_all`.
El payload de un batch contiene, como ya se menciono, una o mas apuestas concatenadas, previa a cada una hay un **header de 4 bytes que avisa sobre la longitud de la misma**:

```
[4 bytes: longitud_apuesta_1] [apuesta_1] [4 bytes: longitud_apuesta_2] [apuesta_2] ... [4 bytes: longitud_apuesta_n] [apuesta_n]
```
Este delimitador de longitud de la apusta nos permite que el receptor pueda separar las apuestas del batch.

Ya despues el mensaje de `END` mas alla del primer byte avisando el tipo de mensaje, los 4 bytes restantes no son relevantes (contienen _basura_).
Se opto por hacer esto ya que sino cada vez que se queria leer el proximo batch/bet se deberia:
- Primero ver el tipo de mensaje.
- Despues la longitud del mismo.
- Leer efectivamente el batch/bet.

De esta manera tendriamos tres llamadas a `recv_all` cada vez que queramos leer una apuesta o un batch. 
Las llamadas a `recv_all` son muy costosas ya que se esta haciendo un cambio de contexto de modo usuario a modo kernel, por lo que la idea es evitarlas lo maximo posible.

Por eso se opto por unir los primeros dos pasos a:
- Leer tipo de mensaje y longitud del mismo.
- Leer batch/bet.

Lo cual implicaria una llamada menos al `recv_all` por cada vez que se quiere leer una batch/bet. El trade off de esto es que para enviar el mensaje `END` se esta _enviando basura_ pero se considero que enviar **4 bytes basura** una vez cada tanto(solo serian dos veces en todo el flujo de la comunicación) es mucho mejor que estar haciendo una llamada mas al `recv_all` por cada vez que se quiere leer una apuesta/batch.

Y finalmente, los mensajes de `ACK_OK`/`ACK_FAIL` ocupan unicamente un byte. 

#### Serializacion de la apuesta
Cada apuesta se serializa combinando campos de tamaño fijo y variable:

- **Campos fijos**: 14 bytes que contienen información sobre  `agency_id`, `document`, `number`,`birthdate`(el cual es un _string_ de 10 bytes en formato `YYYY-MM-DD`).

- **Campos dinamicos**: `first_name`, `last_name`. Como su longitud en bytes puede variar a cada uno se le ponen 2 bytes antes de su contenido que indican cuantos bytes ocupan.

Cabe aclarar tambien, que el protocolo simplemente se encarga de la serializacion/deserializacion y de hacer los envios a traves de la capa de comunicación, por lo que no sabe ni conoce la logica de negocio de `Lottery`, lo unico que sabe sobre la misma es transformar bytes en `Bets` y viceversa.
Esto se hace en esta capa ya que la logica de negocio no tiene que estar acoplada a como se hace el envio de mensajes, es decir, a la `Bet` no le interesa como se "se la transforma" para enviarla, si esta logica se hiciera en la `Bet` (ya sea pidiendole que se serialice o algo por el estilo), se la estaria acoplando a la capa de comunicación y si el dia de mañana se la quiere cambiar habria que tocar la logica de negocio, lo cual no es correcto. 
La idea es que si en algun momento se decide cambiar el protocolo de comunicación no deba tocarse la logica de negocio para poder lograrlo.

#### Procesamiento por batches

Respecto al procesamiento por batches el cliente agrupa las apuestas que lee en `INPUT_FILE` en batches de tamaño `BATCH_SIZE`(la cual es una constante que es configurable a travez de una variable de entorno) y envia ese batch en un **unico mensaje** de tipo `BET`. Luego de este mensaje el cliente se queda esperando a la respuesta de el servidor, el cual puede responder con un `ACK_OK` o `ACK_FAIL`.

La decision de saber cuantas apuestas van en cada batch, sumado con la decision de cuando cerrar uno si es que se llego por ejemplo al final del archivo son completar el `BATCH_SIZE`, es responsabilidad de **la capa de aplicación (o sea `client.go`), no del protocolo, el cual unicamente sabe serializar y enviar la lista de apuestas que se le pasa.

#### Flujo completo

1. Cliente envia uno o mas mensajes de tipo `BET` (uno por cada batch) y en cada uno de estos espera el `ACK` del servidor para continuar con el siguiente.

2. Cuando el cliente termina de leer el `INPUT_FILE` envia el mensaje de tipo `END`.

3. El servidor recibe las apuestas hasta que recibe el `END` de parte del cliente, en ese momento la agencia queda a la espera de que se haga el sorteo.

4. Una vez que el servidor alcanza el _quorum de agencias_, calcula los ganadores de esa agencia y envia a las apuestas ganadoras una por una como mensajes de tipo `BET`, y de igual manera cuando termina envia un mensaje de tipo `END`.

5. El cliente recibe los ganadores hasta que recibe el `END` y los persiste en `OUTPUT_FILE`.

### Mecanismos de sincronizacion de la ejecución concurrente

#### Un hilo por cliente en el servidor
El servidor acepta conexiones en un loop y por cada conexion que le llega, crea un `threading.Thread` que se encarga de manejar a esa agencia (recepción de apuestas, espera del quorum, calcula y envia a los ganadores). Esto permite que se puedan procesar varias agenciaas en paralelo en vez de "atenderlas" de forma serial como era al principio.

Se uso `threading` porque en CPython el GIL impide que dos hilos ejecuten bytecode de python al mismo tiempo. Pero como el GIL se libera en las llamadas bloqueantes de I/O, y el servidor apsa la mayor parte del tiempo esperando datos de al red o escribiendo en el archivo de storage (o sesa no es que esta haciendo cosas de computo pesado de CPU), `threading` permite que mientras un hilo esta bloqueado en I/O otro este avanzando. 
En resumen, como en esta aplicación no esta estan haciendo funcionalidades que necesiten computo puro (lo cual necesitaaria paralelismo real con `multiprocessing`) se opto por usar `threading`.

#### Espera del quorum con una variable de condición
Como el sorteo puede realizarse una vez que el numero minimo de agencias (`AGENCY_QUORUM_MIN`) termino de enviar sus apuestas, se opto por utilizar `threading.Condition` junto con un contador compartido entre hilos (`agencies_finished`).

Entonces, cada hilo del servidor al terminar de recibir las apuestas de su agencia incrementa `agencies_finished` dentro de la sección critica. Si todavia no se llego al quorum, el hilo se bloque con un `wait_for` y libera el lock mientras espera. 
Cuando un hilo alcanza el quorum llama a `notify_all()` y de esta manera despierta a todos los hilos bloqueados para que continuen con el calculo de ganadores.

En este mecanismo **todos los hilos permanecen bloqueados de manera eficiente hasta que la condicion se cumple y son "despertados**.

#### Protección de recursos compartidos
Como tenemos varios hilos que estan atentiendo a las distintas agencias en paralelo, se tuvieron que "proteger" los accesos a algunos recursos compartidos entre estos hilos:

1. El archivo de storage: ya que varios hilos sino podrian escribir de manera simultanea en el CSV de apuestas o leerlo mientras otro estaba escribiendo.

2. El contador de `agencies_finished`: si dos hilos lo incrementarian al mismo tiempo se podrian estar perdiendo incrementos.

3. La señal de quorum alcanzada: los hilos que terminan antes tienen que quedar bloqueados sin consumir CPU, o sea sin polling hasta que el hilo que completa el quorum los despierte.

La solución propuesta para mitigar los puntos 2 y 3 se resolvieron con `quorum_condition` que ya se explico arriba (o sea el incremento del contador esta en una seccion critica).
Y para el punto 1 se uso un lock (`file_lock`).

### Terminación graceful con SIGTERM
Tanto el servidor como el cliente tienen un manejador para la señal `SIGTERM`.

#### Servidor
El servidor al recibir la señal se marca como `running=False`, se despiertan a todos los hilos bloqueados en el quorum(para que no se queden esperando para siempre), se cierra el socket que escucha nuevas conexiones y se cierran todos los sockets de clientes activos(desbloquea cualquier `recv`/`send` en cuerso).
Finalmente, el hilo principal espera con un `join` con un timeout predefinido mediante una constante, a que todos los hilos de los clientes finalicen antes de terminan con el proceso.

#### Cliente
En el cliente se crea un `context.Context` que se cancela con si llega `SIGTERM` (con `signal.NotifyCOntext`). Una gorutine espera que se de la cancelacion del contexto y ciera la conexion TCP asi se desbloquea cualquier `recv`/`send`  (o sea lectura o escritura) bloqueante. 
Ademas, el loop principal tambien revisa el `ctx.Done()` entre las distintas iteraciones para cortar el envio de las apuestas ni bien se posible, liberando asi los archivos y la conexion mediante los `defer`.

De esta forma, cuando llega la _señal de termiancion_, ambos procesos liberan los sockets,hilos y archivos que tenian abiertos en un tiempo acotado sin dejar recursos "colgados" ni conexiones "a medio procesar".