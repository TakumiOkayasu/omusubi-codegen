#ifndef IDEVICE_HPP_
#define IDEVICE_HPP_

#include <cstdint>
#include <cstddef>

namespace omusubi {

class IDevice {
public:
    virtual ~IDevice() = default;

    // Initialize the device
    virtual void initialize() = 0;

    // Read data from device
    virtual int read(uint8_t* buffer, size_t size) = 0;

    // Write data to device
    virtual int write(const uint8_t* buffer, size_t size) = 0;

    // Get device status
    virtual bool isReady() const = 0;
};

} // namespace omusubi

#endif // IDEVICE_HPP_
