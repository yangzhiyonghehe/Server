
echo ��ʼע�����

set DocumentPath=%~dp0

set Filepath=mywebsocket.exe

set mainPahth=%DocumentPath%%Filepath% 

echo %mainPahth%

sc create AllenganceWSSServer   binPath= %mainPahth% 
 
sc Start AllenganceWSSServer   
 
sc config AllenganceWSSServer   start= auto
 
sc config AllenganceWSSServer   DisplayName= "AllenganceWSSServer"
 
sc description AllenganceWSSServer   @wss����
 
sc failure AllenganceWSSServer   reset= 30 actions= restart/60000
 

