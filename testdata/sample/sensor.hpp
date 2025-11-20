#ifndef ISENSOR_HPP_
#define ISENSOR_HPP_

namespace omusubi {

class ISensor {
public:
    virtual ~ISensor() = default;

    // Start sensor measurement
    virtual void start() = 0;

    // Stop sensor measurement
    virtual void stop() = 0;

    // Read sensor value
    virtual float readValue() = 0;
};

} // namespace omusubi

#endif // ISENSOR_HPP_
