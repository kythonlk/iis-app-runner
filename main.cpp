#include <windows.h>
#include <iostream>
#include <string>
#include <vector>
#include <tchar.h>

class ServiceManager {
private:
    SC_HANDLE schSCManager;

public:
    ServiceManager() {
        schSCManager = OpenSCManager(
            NULL,                    // local computer
            NULL,                    // database 
            SC_MANAGER_ALL_ACCESS    // full access 
        );
    }

    ~ServiceManager() {
        if (schSCManager) {
            CloseServiceHandle(schSCManager);
        }
    }

    bool StartService(const std::wstring& serviceName) {
        SC_HANDLE schService = OpenService(
            schSCManager,
            serviceName.c_str(),
            SERVICE_ALL_ACCESS
        );

        if (schService == NULL) {
            return false;
        }

        bool success = ::StartService(schService, 0, NULL);
        CloseServiceHandle(schService);
        return success;
    }

    bool StopService(const std::wstring& serviceName) {
        SC_HANDLE schService = OpenService(
            schSCManager,
            serviceName.c_str(),
            SERVICE_ALL_ACCESS
        );

        if (schService == NULL) {
            return false;
        }

        SERVICE_STATUS_PROCESS ssp;
        DWORD dwBytesNeeded;

        if (!ControlService(
            schService,
            SERVICE_CONTROL_STOP,
            (LPSERVICE_STATUS)&ssp)) {
            CloseServiceHandle(schService);
            return false;
        }

        CloseServiceHandle(schService);
        return true;
    }
};

int main() {
    WSADATA wsaData;
    WSAStartup(MAKEWORD(2, 2), &wsaData);

    HANDLE hPipe = CreateNamedPipe(
        L"\\\\.\\pipe\\ServiceManagerPipe",
        PIPE_ACCESS_DUPLEX,
        PIPE_TYPE_MESSAGE | PIPE_READMODE_MESSAGE | PIPE_WAIT,
        1,
        1024,
        1024,
        0,
        NULL
    );

    ServiceManager manager;
    
    while (true) {
        ConnectNamedPipe(hPipe, NULL);
        
        char buffer[1024];
        DWORD bytesRead;
        
        if (ReadFile(hPipe, buffer, sizeof(buffer), &bytesRead, NULL)) {
            std::string command(buffer, bytesRead);
        }
        
        DisconnectNamedPipe(hPipe);
    }

    WSACleanup();
    return 0;
}
