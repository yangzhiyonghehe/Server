
echo ��ʼע�����

set DocumentPath=%~dp0

set Filepath=web.exe

set mainPahth=%DocumentPath%%Filepath% 

echo %mainPahth%

sc create AllenganceServer   binPath= %mainPahth% 
 
sc Start AllenganceServer   
 
sc config AllenganceServer   start= auto
 
sc config AllenganceServer   DisplayName= "AllenganceServer"
 
sc description AllenganceServer   @���ڷ���
 
sc failure AllenganceServer   reset= 30 actions= restart/60000
 
